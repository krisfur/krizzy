// Initialize SortableJS for drag and drop functionality
document.addEventListener('DOMContentLoaded', function() {
    initializeSortable();
});

// Re-initialize after HTMX swaps content
document.addEventListener('htmx:afterSwap', function() {
    initializeSortable();
});

function getBoardId() {
    const container = document.getElementById('columns-container');
    return container ? container.dataset.boardId : null;
}

function initializeSortable() {
    // Make columns sortable
    const columnsContainer = document.getElementById('columns-container');
    if (columnsContainer && !columnsContainer._sortable) {
        columnsContainer._sortable = new Sortable(columnsContainer, {
            animation: 150,
            draggable: '[data-column-id]',
            ghostClass: 'sortable-ghost',
            chosenClass: 'sortable-chosen',
            handle: '.column-header',
            onEnd: function(evt) {
                const boardId = columnsContainer.dataset.boardId;
                const columnIds = Array.from(columnsContainer.querySelectorAll('[data-column-id]'))
                    .map(col => col.dataset.columnId);

                const formData = new FormData();
                formData.append('board_id', boardId);
                columnIds.forEach(id => formData.append('column_ids', id));

                fetch('/columns/reorder', {
                    method: 'POST',
                    body: formData
                });
            }
        });
    }

    // Make cards sortable within and between columns
    const cardContainers = document.querySelectorAll('.cards-container');
    cardContainers.forEach(container => {
        if (!container._sortable) {
            container._sortable = new Sortable(container, {
                group: 'cards',
                animation: 150,
                draggable: '.card-item',
                ghostClass: 'sortable-ghost',
                chosenClass: 'sortable-chosen',
                onEnd: function(evt) {
                    const cardId = evt.item.dataset.cardId;
                    const newColumnId = evt.to.dataset.columnId;
                    const newPosition = evt.newIndex;
                    const boardId = getBoardId();

                    fetch(`/cards/${cardId}/move`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            column_id: parseInt(newColumnId),
                            position: newPosition,
                            board_id: parseInt(boardId)
                        })
                    }).then(response => {
                        if (response.ok && boardId) {
                            // Refresh the board to show updated completion status
                            htmx.ajax('GET', `/boards/${boardId}`, {target: '#board-content', swap: 'innerHTML'});
                        }
                    });
                }
            });
        }
    });

    // Make checklist items sortable
    const checklistContainers = document.querySelectorAll('.checklist-container');
    checklistContainers.forEach(container => {
        if (!container._sortable) {
            container._sortable = new Sortable(container, {
                animation: 150,
                draggable: '.checklist-item',
                ghostClass: 'sortable-ghost',
                chosenClass: 'sortable-chosen',
                filter: 'input, button',
                onEnd: function(evt) {
                    const cardId = container.dataset.cardId;
                    const itemIds = Array.from(container.querySelectorAll('.checklist-item'))
                        .map(item => item.dataset.itemId);

                    const boardId = getBoardId();
                    const formData = new FormData();
                    itemIds.forEach(id => formData.append('item_ids', id));
                    if (boardId) formData.append('board_id', boardId);

                    fetch(`/cards/${cardId}/checklist/reorder`, {
                        method: 'POST',
                        body: formData
                    });
                }
            });
        }
    });
}

// Refresh board when modal closes
function closeModalAndRefresh(boardId) {
    document.getElementById('modal-backdrop').classList.add('hidden');
    if (!boardId) {
        boardId = getBoardId();
    }
    if (boardId) {
        htmx.ajax('GET', `/boards/${boardId}`, {target: '#board-content', swap: 'innerHTML'});
    }
}

// Board rename functions
function startRenameBoard(boardId, currentName) {
    const form = document.getElementById(`rename-form-${boardId}`);
    const input = document.getElementById(`rename-input-${boardId}`);
    if (form) {
        form.classList.remove('hidden');
        input.value = currentName;
        input.focus();
        input.select();
    }
}

function cancelRenameBoard(boardId) {
    const form = document.getElementById(`rename-form-${boardId}`);
    if (form) {
        form.classList.add('hidden');
    }
}
