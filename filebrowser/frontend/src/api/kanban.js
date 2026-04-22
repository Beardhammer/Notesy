import { fetchJSON, fetchURL } from "./utils";

// Boards
export async function getBoards() {
  return fetchJSON("/api/boards", {});
}

export async function createBoard(name) {
  return fetchJSON("/api/boards", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export async function updateBoard(id, name) {
  return fetchJSON(`/api/boards/${id}`, {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
}

export async function deleteBoard(id) {
  await fetchURL(`/api/boards/${id}`, { method: "DELETE" });
}

// Board-scoped Tasks
export async function getTasks(boardId) {
  if (boardId === "all") {
    return fetchJSON("/api/boards/all/kanban", {});
  }
  return fetchJSON(`/api/boards/${boardId}/kanban`, {});
}

export async function getTask(boardId, id) {
  return fetchJSON(`/api/boards/${boardId}/kanban/${id}`, {});
}

export async function createTask(boardId, task) {
  return fetchJSON(`/api/boards/${boardId}/kanban`, {
    method: "POST",
    body: JSON.stringify(task),
  });
}

export async function updateTask(boardId, id, task) {
  // When in "all" view, use the task's own boardId
  const bid = boardId === "all" ? task.boardId : boardId;
  return fetchJSON(`/api/boards/${bid}/kanban/${id}`, {
    method: "PUT",
    body: JSON.stringify(task),
  });
}

export async function deleteTask(boardId, id) {
  await fetchURL(`/api/boards/${boardId}/kanban/${id}`, { method: "DELETE" });
}

// Board-scoped Events
export async function getEvents(boardId, from, to) {
  if (boardId === "all") {
    return fetchJSON(`/api/boards/all/events?from=${from}&to=${to}`, {});
  }
  return fetchJSON(`/api/boards/${boardId}/events?from=${from}&to=${to}`, {});
}

export async function getAllEvents(boardId) {
  if (boardId === "all") {
    return fetchJSON("/api/boards/all/events", {});
  }
  return fetchJSON(`/api/boards/${boardId}/events`, {});
}

export async function createEvent(boardId, event) {
  return fetchJSON(`/api/boards/${boardId}/events`, {
    method: "POST",
    body: JSON.stringify(event),
  });
}

export async function updateEvent(boardId, id, event) {
  const bid = boardId === "all" ? event.boardId : boardId;
  return fetchJSON(`/api/boards/${bid}/events/${id}`, {
    method: "PUT",
    body: JSON.stringify(event),
  });
}

export async function deleteEvent(boardId, id) {
  await fetchURL(`/api/boards/${boardId}/events/${id}`, { method: "DELETE" });
}
