package services

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"krizzy/internal/models"
)

type TrelloImportService struct {
	bm *BoardManager
}

func NewTrelloImportService(bm *BoardManager) *TrelloImportService {
	return &TrelloImportService{bm: bm}
}

type trelloBoardExport struct {
	Name       string            `json:"name"`
	Lists      []trelloList      `json:"lists"`
	Cards      []trelloCard      `json:"cards"`
	Checklists []trelloChecklist `json:"checklists"`
	Members    []trelloMember    `json:"members"`
}

type trelloList struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Pos    float64 `json:"pos"`
	Closed bool    `json:"closed"`
}

type trelloCard struct {
	ID           string   `json:"id"`
	IDList       string   `json:"idList"`
	Name         string   `json:"name"`
	Desc         string   `json:"desc"`
	Pos          float64  `json:"pos"`
	Closed       bool     `json:"closed"`
	IDMembers    []string `json:"idMembers"`
	IDChecklists []string `json:"idChecklists"`
}

type trelloChecklist struct {
	ID         string            `json:"id"`
	IDCard     string            `json:"idCard"`
	Name       string            `json:"name"`
	Pos        float64           `json:"pos"`
	CheckItems []trelloCheckItem `json:"checkItems"`
}

type trelloCheckItem struct {
	Name  string  `json:"name"`
	Pos   float64 `json:"pos"`
	State string  `json:"state"`
}

type trelloMember struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Username string `json:"username"`
}

func (s *TrelloImportService) ImportBoard(r io.Reader, boardName, dbType string, pgConnectionID *int64, pgDatabaseName string) (*models.Board, error) {
	var export trelloBoardExport
	if err := json.NewDecoder(r).Decode(&export); err != nil {
		return nil, fmt.Errorf("failed to parse Trello export: %w", err)
	}

	importName := strings.TrimSpace(boardName)
	if importName == "" {
		importName = strings.TrimSpace(export.Name)
	}
	if importName == "" {
		return nil, fmt.Errorf("board name is required")
	}

	activeLists := make([]trelloList, 0, len(export.Lists))
	activeListIDs := make(map[string]struct{}, len(export.Lists))
	for _, list := range export.Lists {
		if list.Closed || isArchiveList(list.Name) {
			continue
		}
		activeLists = append(activeLists, list)
		activeListIDs[list.ID] = struct{}{}
	}
	if len(activeLists) == 0 {
		return nil, fmt.Errorf("no non-archive lists found in Trello export")
	}
	sort.Slice(activeLists, func(i, j int) bool {
		return activeLists[i].Pos < activeLists[j].Pos
	})

	activeCards := make([]trelloCard, 0, len(export.Cards))
	usedMemberIDs := make(map[string]struct{})
	for _, card := range export.Cards {
		if card.Closed {
			continue
		}
		if _, ok := activeListIDs[card.IDList]; !ok {
			continue
		}
		activeCards = append(activeCards, card)
		for _, memberID := range card.IDMembers {
			usedMemberIDs[memberID] = struct{}{}
		}
	}

	memberNames := make(map[string]string, len(export.Members))
	usedNames := make(map[string]struct{}, len(export.Members))
	for _, member := range export.Members {
		name := strings.TrimSpace(member.FullName)
		if name == "" {
			name = strings.TrimSpace(member.Username)
		}
		if name == "" {
			continue
		}
		if _, exists := usedNames[name]; exists {
			username := strings.TrimSpace(member.Username)
			if username != "" {
				name = fmt.Sprintf("%s (@%s)", name, username)
			}
		}
		memberNames[member.ID] = name
		usedNames[name] = struct{}{}
	}

	board, err := s.bm.CreateBoardWithoutDefaults(importName, dbType, pgConnectionID, pgDatabaseName)
	if err != nil {
		return nil, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = s.bm.DeleteBoard(board.ID)
		}
	}()

	svc, err := s.bm.GetServiceForBoard(board.ID)
	if err != nil {
		return nil, err
	}

	personIDsByTrelloID := make(map[string]int64, len(usedMemberIDs))
	memberIDList := make([]string, 0, len(usedMemberIDs))
	for memberID := range usedMemberIDs {
		memberIDList = append(memberIDList, memberID)
	}
	sort.Slice(memberIDList, func(i, j int) bool {
		return memberNames[memberIDList[i]] < memberNames[memberIDList[j]]
	})
	for _, memberID := range memberIDList {
		name := memberNames[memberID]
		if name == "" {
			continue
		}
		person := &models.Person{
			BoardID: board.ID,
			Name:    name,
		}
		if err := svc.PersonRepo.Create(person); err != nil {
			return nil, fmt.Errorf("failed to import person %q: %w", name, err)
		}
		personIDsByTrelloID[memberID] = person.ID
	}

	checklistsByCardID := make(map[string][]trelloChecklist)
	for _, checklist := range export.Checklists {
		checklistsByCardID[checklist.IDCard] = append(checklistsByCardID[checklist.IDCard], checklist)
	}
	for cardID := range checklistsByCardID {
		sort.Slice(checklistsByCardID[cardID], func(i, j int) bool {
			return checklistsByCardID[cardID][i].Pos < checklistsByCardID[cardID][j].Pos
		})
	}

	cardsByListID := make(map[string][]trelloCard)
	for _, card := range activeCards {
		cardsByListID[card.IDList] = append(cardsByListID[card.IDList], card)
	}
	for listID := range cardsByListID {
		sort.Slice(cardsByListID[listID], func(i, j int) bool {
			return cardsByListID[listID][i].Pos < cardsByListID[listID][j].Pos
		})
	}

	now := time.Now()
	for _, list := range activeLists {
		column := &models.Column{
			BoardID:      board.ID,
			Name:         strings.TrimSpace(list.Name),
			IsDoneColumn: strings.EqualFold(strings.TrimSpace(list.Name), "done"),
		}
		if column.Name == "" {
			column.Name = "Untitled"
		}
		if err := svc.ColumnRepo.Create(column); err != nil {
			return nil, fmt.Errorf("failed to import list %q: %w", list.Name, err)
		}

		for _, cardData := range cardsByListID[list.ID] {
			card := &models.Card{
				ColumnID:    column.ID,
				Title:       strings.TrimSpace(cardData.Name),
				Description: cardData.Desc,
			}
			if card.Title == "" {
				card.Title = "Untitled"
			}
			if err := svc.CardRepo.Create(card); err != nil {
				return nil, fmt.Errorf("failed to import card %q: %w", card.Title, err)
			}

			if column.IsDoneColumn {
				card.CompletedAt = &now
				if err := svc.CardRepo.Update(card); err != nil {
					return nil, fmt.Errorf("failed to mark imported done card %q: %w", card.Title, err)
				}
			}

			if err := s.importChecklistItems(svc, card.ID, checklistsByCardID[cardData.ID]); err != nil {
				return nil, fmt.Errorf("failed to import checklist for %q: %w", card.Title, err)
			}

			assigneeIDs := make([]int64, 0, len(cardData.IDMembers))
			for _, memberID := range cardData.IDMembers {
				personID, ok := personIDsByTrelloID[memberID]
				if ok {
					assigneeIDs = append(assigneeIDs, personID)
				}
			}
			if len(assigneeIDs) > 0 {
				if err := svc.PersonRepo.SetCardAssignees(card.ID, assigneeIDs); err != nil {
					return nil, fmt.Errorf("failed to import assignees for %q: %w", card.Title, err)
				}
			}
		}
	}

	cleanup = false
	return board, nil
}

func (s *TrelloImportService) importChecklistItems(svc *KanbanService, cardID int64, checklists []trelloChecklist) error {
	if len(checklists) == 0 {
		return nil
	}

	multipleChecklists := len(checklists) > 1
	for _, checklist := range checklists {
		items := append([]trelloCheckItem(nil), checklist.CheckItems...)
		sort.Slice(items, func(i, j int) bool {
			return items[i].Pos < items[j].Pos
		})

		for _, item := range items {
			content := strings.TrimSpace(item.Name)
			if content == "" {
				continue
			}
			if multipleChecklists {
				checklistName := strings.TrimSpace(checklist.Name)
				if checklistName != "" {
					content = checklistName + ": " + content
				}
			}

			checklistItem := &models.ChecklistItem{
				CardID:      cardID,
				Content:     content,
				IsCompleted: item.State == "complete",
			}
			if err := svc.ChecklistRepo.Create(checklistItem); err != nil {
				return err
			}
		}
	}

	return nil
}

func isArchiveList(name string) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(name)), "archive")
}
