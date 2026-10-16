import { onChange, loadBoards } from "./boards.js";

const res = await fetch("/leaderboards");
const boards = await res.json();

loadBoards(boards)

onChange(async ({ boards }) => {
    const res = await fetch("/leaderboards", {
        method: "PUT",
        body: JSON.stringify(boards)
    });

    console.log(res.status, res.statusText)
})
