const $ = (id) => document.getElementById(id);
const esc = (v) =>
    String(v == null ? "" : v).replace(
        /[&<>"']/g,
        (c) =>
            ({
                "&": "&amp;",
                "<": "&lt;",
                ">": "&gt;",
                '"': "&quot;",
                "'": "&#39;",
            })[c],
    );

const els = {
    boardForm: $("new-board-form"),
    boardName: $("new-board-name"),
    boardError: $("new-board-error"),
    boardList: $("board-list"),
    title: $("board-title"),
    subtitle: $("board-subtitle"),
    deleteBoard: $("delete-board"),
    playerForm: $("add-player-form"),
    playerName: $("player-name"),
    playerError: $("player-error"),
    playerList: $("player-list"),
};

let boards = {};
let activeName = null;
const listeners = new Set();

export function getBoards() {
    return boards;
}

export function getActiveBoard() {
    return activeName;
}

export function onChange(listener) {
    listeners.add(listener);
    return () => listeners.delete(listener);
}

function notify(name) {
    const players = name != null && name in boards ? boards[name] : null;
    for (const listener of listeners) {
        listener({ name, players, boards });
    }
}

function toEntries(data) {
    if (data instanceof Map) return [...data];
    if (Array.isArray(data)) return data;
    if (data && typeof data === "object") return Object.entries(data);
    return [];
}

export function loadBoards(data) {
    boards = {};
    for (const [name, players] of toEntries(data)) {
        boards[String(name)] = Array.isArray(players) ? [...players] : [];
    }
    activeName =
        activeName != null && activeName in boards
            ? activeName
            : (Object.keys(boards)[0] ?? null);
    showError(els.boardError, "");
    showError(els.playerError, "");
    render();
}

function activeBoard() {
    return activeName != null ? boards[activeName] : undefined;
}

function showError(el, msg) {
    el.hidden = !msg;
    el.textContent = msg || "";
}

function playerCount(n) {
    return `${n} ${n === 1 ? "player" : "players"}`;
}

function renderBoards() {
    const entries = Object.entries(boards);
    els.boardList.innerHTML = entries.length
        ? entries
              .map(
                  ([name, players]) => `<li class="board-item${name === activeName ? " active" : ""}">
        <button type="button" class="board-select" data-name="${esc(name)}">
          <strong>${esc(name)}</strong>
          <span>${playerCount(players.length)}</span>
        </button>
        <button type="button" class="icon-btn danger" data-del="${esc(name)}" title="Delete leaderboard" aria-label="Delete ${esc(name)}">✕</button>
      </li>`,
              )
              .join("")
        : '<li class="empty">No leaderboards yet.</li>';
}

function renderBoard() {
    const players = activeBoard();

    els.deleteBoard.hidden = !players;
    els.playerForm.hidden = !players;
    els.title.textContent = activeName ?? "No leaderboard selected";
    els.subtitle.textContent = players
        ? `${playerCount(players.length)} · ranked by order added`
        : "Create one on the left to get started";

    if (!players) {
        els.playerList.innerHTML = '<li class="empty">No players yet.</li>';
        return;
    }

    els.playerList.innerHTML = players.length
        ? players
              .map(
                  (name, i) => `<li class="player-row">
        <span class="rank">#${i + 1}</span>
        <strong class="player-name">${esc(name)}</strong>
        <span class="row-actions">
          <button type="button" class="icon-btn" data-move="up" data-i="${i}" title="Move up" ${i === 0 ? "disabled" : ""}>↑</button>
          <button type="button" class="icon-btn" data-move="down" data-i="${i}" title="Move down" ${i === players.length - 1 ? "disabled" : ""}>↓</button>
          <button type="button" class="icon-btn danger" data-remove="${i}" title="Remove player" aria-label="Remove ${esc(name)}">✕</button>
        </span>
      </li>`,
              )
              .join("")
        : '<li class="empty">No players yet — add names above.</li>';
}

function render() {
    renderBoards();
    renderBoard();
}

els.boardForm.addEventListener("submit", (event) => {
    event.preventDefault();
    const name = els.boardName.value.trim();
    if (!name) return;
    if (
        Object.keys(boards).some((k) => k.toLowerCase() === name.toLowerCase())
    ) {
        showError(els.boardError, "A leaderboard with that name already exists.");
        return;
    }
    showError(els.boardError, "");

    boards[name] = [];
    activeName = name;
    els.boardName.value = "";
    render();
    notify(name);
    els.playerName.focus();
});

els.boardList.addEventListener("click", (event) => {
    const del = event.target.closest("[data-del]");
    if (del) {
        const name = del.dataset.del;
        delete boards[name];
        if (activeName === name)
            activeName = Object.keys(boards)[0] ?? null;
        showError(els.playerError, "");
        render();
        notify(name);
        return;
    }

    const select = event.target.closest("[data-name]");
    if (select) {
        activeName = select.dataset.name;
        showError(els.playerError, "");
        render();
        els.playerName.focus();
    }
});

els.deleteBoard.addEventListener("click", () => {
    if (activeName == null || !(activeName in boards)) return;
    const name = activeName;
    delete boards[name];
    activeName = Object.keys(boards)[0] ?? null;
    showError(els.playerError, "");
    render();
    notify(name);
});

els.playerForm.addEventListener("submit", (event) => {
    event.preventDefault();
    const name = activeName;
    const players = activeBoard();
    if (!players) return;

    const player = els.playerName.value.trim();
    if (!player) return;
    if (players.some((p) => p.toLowerCase() === player.toLowerCase())) {
        showError(els.playerError, "That player is already on this leaderboard.");
        return;
    }
    showError(els.playerError, "");

    players.push(player);
    els.playerName.value = "";
    render();
    notify(name);
    els.playerName.focus();
});

els.playerList.addEventListener("click", (event) => {
    const name = activeName;
    const players = activeBoard();
    if (!players) return;

    const remove = event.target.closest("[data-remove]");
    if (remove) {
        players.splice(Number(remove.dataset.remove), 1);
        render();
        notify(name);
        return;
    }

    const move = event.target.closest("[data-move]");
    if (move) {
        const i = Number(move.dataset.i);
        const j = move.dataset.move === "up" ? i - 1 : i + 1;
        if (j < 0 || j >= players.length) return;
        [players[i], players[j]] = [players[j], players[i]];
        render();
        notify(name);
    }
});

render();
