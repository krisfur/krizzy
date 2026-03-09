var krizzyRealtime = {
    clientId: null,
    boardId: null,
    eventSource: null
};

function getClientId() {
    if (krizzyRealtime.clientId) {
        return krizzyRealtime.clientId;
    }

    var storedId = window.sessionStorage.getItem('krizzy-client-id');
    if (storedId) {
        krizzyRealtime.clientId = storedId;
        return storedId;
    }

    var newId = 'client-' + Math.random().toString(36).slice(2) + Date.now().toString(36);
    window.sessionStorage.setItem('krizzy-client-id', newId);
    krizzyRealtime.clientId = newId;
    return newId;
}

function getBoardId() {
    var container = document.getElementById('columns-container');
    return container ? container.dataset.boardId : null;
}

function getCurrentModalCardId() {
    var modalCard = document.querySelector('#modal-content [data-card-id]');
    return modalCard ? modalCard.dataset.cardId : null;
}

function isPeopleModalOpen(boardId) {
    return !!document.getElementById('people-list') && !!document.querySelector('#modal-content [data-board-id="' + boardId + '"]');
}

function withClientHeaders(headers) {
    var merged = headers || {};
    merged['X-Client-ID'] = getClientId();
    return merged;
}

function refreshColumnsContainer(boardId) {
    htmx.ajax('GET', '/boards/' + boardId + '/columns', {
        target: '#columns-container',
        swap: 'outerHTML'
    });
}

function refreshColumn(boardId, columnId) {
    var target = document.getElementById('column-' + columnId);
    if (!target) {
        refreshColumnsContainer(boardId);
        return;
    }

    htmx.ajax('GET', '/boards/' + boardId + '/columns/' + columnId, {
        target: '#column-' + columnId,
        swap: 'outerHTML'
    });
}

function refreshCard(boardId, cardId, fallbackColumnId) {
    var target = document.getElementById('card-' + cardId);
    if (!target) {
        if (fallbackColumnId) {
            refreshColumn(boardId, fallbackColumnId);
        }
        return;
    }

    htmx.ajax('GET', '/boards/' + boardId + '/cards/' + cardId, {
        target: '#card-' + cardId,
        swap: 'outerHTML'
    });
}

function refreshOpenCardModal(boardId, cardId) {
    if (getCurrentModalCardId() !== String(cardId)) {
        return;
    }

    htmx.ajax('GET', '/cards/' + cardId + '/modal?board_id=' + boardId, {
        target: '#modal-content',
        swap: 'innerHTML'
    });
}

function refreshPeopleModal(boardId) {
    if (!isPeopleModalOpen(String(boardId))) {
        return;
    }

    htmx.ajax('GET', '/boards/' + boardId + '/people', {
        target: '#modal-content',
        swap: 'innerHTML'
    });
}

function handleBoardEvent(event) {
    var boardId = getBoardId();
    if (!boardId || String(event.board_id) !== String(boardId)) {
        return;
    }

    if (event.client_id && event.client_id === getClientId()) {
        return;
    }

    switch (event.type) {
        case 'column.created':
        case 'column.deleted':
        case 'column.reordered':
            refreshColumnsContainer(boardId);
            break;
        case 'column.updated':
            if (event.column_id) {
                refreshColumn(boardId, event.column_id);
            }
            break;
        case 'card.created':
        case 'card.deleted':
            if (event.column_id) {
                refreshColumn(boardId, event.column_id);
            }
            break;
        case 'card.updated':
            if (event.card_id) {
                refreshCard(boardId, event.card_id, event.column_id);
                refreshOpenCardModal(boardId, event.card_id);
            }
            break;
        case 'card.moved':
            if (event.from_column_id) {
                refreshColumn(boardId, event.from_column_id);
            }
            if (event.to_column_id && event.to_column_id !== event.from_column_id) {
                refreshColumn(boardId, event.to_column_id);
            }
            if (event.card_id) {
                refreshOpenCardModal(boardId, event.card_id);
            }
            break;
        case 'checklist.updated':
            if (event.card_id) {
                refreshCard(boardId, event.card_id, event.column_id);
                refreshOpenCardModal(boardId, event.card_id);
            }
            break;
        case 'comment.updated':
            if (event.card_id) {
                refreshOpenCardModal(boardId, event.card_id);
            }
            break;
        case 'people.updated':
            refreshPeopleModal(boardId);
            refreshColumnsContainer(boardId);
            if (getCurrentModalCardId()) {
                refreshOpenCardModal(boardId, getCurrentModalCardId());
            }
            break;
    }
}

function initializeRealtime() {
    var boardId = getBoardId();

    if (!boardId) {
        if (krizzyRealtime.eventSource) {
            krizzyRealtime.eventSource.close();
            krizzyRealtime.eventSource = null;
        }
        krizzyRealtime.boardId = null;
        return;
    }

    if (krizzyRealtime.eventSource && krizzyRealtime.boardId === boardId) {
        return;
    }

    if (krizzyRealtime.eventSource) {
        krizzyRealtime.eventSource.close();
    }

    krizzyRealtime.boardId = boardId;
    krizzyRealtime.eventSource = new EventSource('/boards/' + boardId + '/events');
    krizzyRealtime.eventSource.addEventListener('board-update', function(message) {
        handleBoardEvent(JSON.parse(message.data));
    });
    krizzyRealtime.eventSource.onerror = function() {
        if (!getBoardId() && krizzyRealtime.eventSource) {
            krizzyRealtime.eventSource.close();
            krizzyRealtime.eventSource = null;
            krizzyRealtime.boardId = null;
        }
    };
}

// Show HTMX error responses as alerts
document.addEventListener('htmx:responseError', function(event) {
    var elt = event.detail.elt;
    if (elt && (elt.id === 'trello-import-form' || elt.closest('#trello-import-form'))) {
        showImportFeedback('error', event.detail.xhr && event.detail.xhr.responseText ? event.detail.xhr.responseText : 'Import failed.');
        setImportFormDisabled(false);
        return;
    }

    var xhr = event.detail.xhr;
    if (xhr && xhr.responseText) {
        alert(xhr.responseText);
    }
});

document.addEventListener('htmx:configRequest', function(event) {
    event.detail.headers['X-Client-ID'] = getClientId();
});

// Initialize SortableJS and realtime hooks
document.addEventListener('DOMContentLoaded', function() {
    initializeSortable();
    initializeRealtime();

    toggleCreatePgFields();
    toggleImportPgFields();
});

// Re-initialize after HTMX swaps content
document.addEventListener('htmx:afterSwap', function() {
    initializeSortable();
    initializeRealtime();
    toggleCreatePgFields();
    toggleImportPgFields();
});

function toggleCreatePgFields() {
    var dbTypeSelect = document.querySelector('#boards-list select[name="db_type"]');
    var pgFields = document.getElementById('pg-fields');
    if (dbTypeSelect && pgFields) {
        pgFields.style.display = dbTypeSelect.value === 'postgres' ? 'block' : 'none';
    }
}

function toggleImportPgFields() {
    var dbTypeSelect = document.getElementById('import-db-type');
    var pgFields = document.getElementById('import-pg-fields');
    if (dbTypeSelect && pgFields) {
        pgFields.style.display = dbTypeSelect.value === 'postgres' ? 'block' : 'none';
    }
}

function openImportModal() {
    var backdrop = document.getElementById('import-modal-backdrop');
    if (!backdrop) {
        return;
    }

    backdrop.classList.remove('hidden');
    htmx.ajax('GET', '/boards/import-modal', {
        target: '#import-modal-content',
        swap: 'innerHTML'
    });
}

function closeImportModal() {
    var backdrop = document.getElementById('import-modal-backdrop');
    if (backdrop) {
        backdrop.classList.add('hidden');
    }
}

function setImportFormDisabled(disabled) {
    var fields = document.getElementById('import-form-fields');
    if (!fields) {
        return;
    }

    Array.from(fields.querySelectorAll('input, select, textarea, button')).forEach(function(el) {
        if (disabled) {
            el.setAttribute('disabled', 'disabled');
        } else if (el.type !== 'button' || el.textContent.trim() !== 'Close') {
            el.removeAttribute('disabled');
        }
    });
    fields.style.opacity = disabled ? '0.55' : '1';
}

function showImportFeedback(state, message) {
    var feedback = document.getElementById('import-feedback');
    if (!feedback) {
        return;
    }

    feedback.classList.remove('hidden', 'border-red-700', 'bg-red-950', 'text-red-200', 'border-green-700', 'bg-green-950', 'text-green-200', 'border-go-blue', 'bg-dark-700', 'text-dark-100');

    if (state === 'loading') {
        feedback.classList.add('border-go-blue', 'bg-dark-700', 'text-dark-100');
        feedback.innerHTML = '<div class="flex items-center gap-3"><span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-go-blue border-t-transparent"></span><div><div class="font-medium">Importing...</div><div class="text-sm text-dark-300">This can take a moment for larger boards.</div></div></div>';
        return;
    }

    if (state === 'success') {
        feedback.classList.add('border-green-700', 'bg-green-950', 'text-green-200');
        feedback.innerHTML = '<div class="font-medium">Success!</div><div class="text-sm mt-1">' + escapeHtml(message || 'Board imported successfully.') + '</div>';
        return;
    }

    feedback.classList.add('border-red-700', 'bg-red-950', 'text-red-200');
    feedback.innerHTML = '<div class="font-medium">Import failed</div><div class="text-sm mt-1">' + escapeHtml(message || 'Something went wrong while importing.') + '</div>';
}

function startImportFeedback() {
    showImportFeedback('loading');
    setImportFormDisabled(true);
}

function finishImportFeedback(event) {
    if (!event.detail.successful) {
        return;
    }

    var form = document.getElementById('trello-import-form');
    if (form) {
        form.reset();
    }
    setImportFormDisabled(false);
    toggleImportPgFields();
    showImportFeedback('success', 'Your Trello board is ready.');
}

function escapeHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

window.toggleCreatePgFields = toggleCreatePgFields;
window.toggleImportPgFields = toggleImportPgFields;
window.openImportModal = openImportModal;
window.closeImportModal = closeImportModal;
window.startImportFeedback = startImportFeedback;
window.finishImportFeedback = finishImportFeedback;

function initializeSortable() {
    var columnsContainer = document.getElementById('columns-container');
    if (columnsContainer && !columnsContainer._sortable) {
        columnsContainer._sortable = new Sortable(columnsContainer, {
            animation: 150,
            draggable: '[data-column-id]',
            ghostClass: 'sortable-ghost',
            chosenClass: 'sortable-chosen',
            handle: '.column-header',
            onEnd: function() {
                var boardId = columnsContainer.dataset.boardId;
                var columnIds = Array.from(columnsContainer.querySelectorAll('[data-column-id]')).map(function(col) {
                    return col.dataset.columnId;
                });

                var formData = new FormData();
                formData.append('board_id', boardId);
                columnIds.forEach(function(id) {
                    formData.append('column_ids', id);
                });

                fetch('/columns/reorder', {
                    method: 'POST',
                    headers: withClientHeaders(),
                    body: formData
                });
            }
        });
    }

    var cardContainers = document.querySelectorAll('.cards-container');
    cardContainers.forEach(function(container) {
        if (!container._sortable) {
            container._sortable = new Sortable(container, {
                group: 'cards',
                animation: 150,
                draggable: '.card-item',
                ghostClass: 'sortable-ghost',
                chosenClass: 'sortable-chosen',
                onEnd: function(evt) {
                    var cardId = evt.item.dataset.cardId;
                    var newColumnId = evt.to.dataset.columnId;
                    var newPosition = evt.newIndex;
                    var boardId = getBoardId();

                    fetch('/cards/' + cardId + '/move', {
                        method: 'POST',
                        headers: withClientHeaders({
                            'Content-Type': 'application/json'
                        }),
                        body: JSON.stringify({
                            column_id: parseInt(newColumnId, 10),
                            position: newPosition,
                            board_id: parseInt(boardId, 10)
                        })
                    }).then(function(response) {
                        if (!boardId) {
                            return;
                        }

                        var fromColumnId = evt.from.dataset.columnId;
                        if (!response.ok) {
                            refreshColumnsContainer(boardId);
                            return;
                        }

                        if (fromColumnId) {
                            refreshColumn(boardId, fromColumnId);
                        }
                        if (newColumnId && newColumnId !== fromColumnId) {
                            refreshColumn(boardId, newColumnId);
                        }
                    });
                }
            });
        }
    });

    var checklistContainers = document.querySelectorAll('.checklist-container');
    checklistContainers.forEach(function(container) {
        if (!container._sortable) {
            container._sortable = new Sortable(container, {
                animation: 150,
                draggable: '.checklist-item',
                ghostClass: 'sortable-ghost',
                chosenClass: 'sortable-chosen',
                filter: 'input, button',
                onEnd: function() {
                    var cardId = container.dataset.cardId;
                    var itemIds = Array.from(container.querySelectorAll('.checklist-item')).map(function(item) {
                        return item.dataset.itemId;
                    });

                    var boardId = getBoardId();
                    var formData = new FormData();
                    itemIds.forEach(function(id) {
                        formData.append('item_ids', id);
                    });
                    if (boardId) {
                        formData.append('board_id', boardId);
                    }

                    fetch('/cards/' + cardId + '/checklist/reorder', {
                        method: 'POST',
                        headers: withClientHeaders(),
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
        refreshColumnsContainer(boardId);
    }
}

// Close connections modal and refresh the boards list (to update connection dropdown)
function closeConnModal() {
    document.getElementById('conn-modal-backdrop').classList.add('hidden');
    htmx.ajax('GET', '/', {target: '#boards-list', swap: 'innerHTML'});
    if (!document.getElementById('import-modal-backdrop').classList.contains('hidden')) {
        htmx.ajax('GET', '/boards/import-modal', {target: '#import-modal-content', swap: 'innerHTML'});
    }
}

// Board rename functions
function startRenameBoard(boardId, currentName) {
    var form = document.getElementById('rename-form-' + boardId);
    var input = document.getElementById('rename-input-' + boardId);
    if (form) {
        form.classList.remove('hidden');
        input.value = currentName;
        input.focus();
        input.select();
    }
}

function cancelRenameBoard(boardId) {
    var form = document.getElementById('rename-form-' + boardId);
    if (form) {
        form.classList.add('hidden');
    }
}

function startRenameColumn(columnId) {
    var form = document.getElementById('rename-column-form-' + columnId);
    var header = document.getElementById('column-header-' + columnId);
    var input = document.getElementById('rename-column-input-' + columnId);
    if (form && input) {
        closeAllColumnRenameForms(columnId);
        form.classList.remove('hidden');
        if (header) {
            header.classList.add('hidden');
        }
        input.focus();
        input.select();
    }
}

function cancelRenameColumn(columnId) {
    var form = document.getElementById('rename-column-form-' + columnId);
    var header = document.getElementById('column-header-' + columnId);
    if (form) {
        form.classList.add('hidden');
    }
    if (header) {
        header.classList.remove('hidden');
    }
}

function closeAllColumnRenameForms(exceptColumnId) {
    document.querySelectorAll('[id^="rename-column-form-"]').forEach(function(form) {
        var id = form.id.replace('rename-column-form-', '');
        if (!exceptColumnId || String(id) !== String(exceptColumnId)) {
            cancelRenameColumn(id);
        }
    });
}

document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        closeAllColumnRenameForms();
    }
});

document.addEventListener('click', function(event) {
    document.querySelectorAll('[id^="rename-column-form-"]').forEach(function(form) {
        if (form.classList.contains('hidden')) {
            return;
        }

        var id = form.id.replace('rename-column-form-', '');
        var renameButton = document.querySelector('[onclick*="startRenameColumn(' + id + ')"]');
        if (form.contains(event.target) || (renameButton && renameButton.contains(event.target))) {
            return;
        }

        cancelRenameColumn(id);
    });
});
