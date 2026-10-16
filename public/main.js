import { setLeaderboard, setBeatmap } from "./dom.js";
import init, { Beatmap, Difficulty, HitResultPriority, Performance } from './rosu/rosu_pp_js.js';

await init();

const ws = new WebSocket(`ws://${location.host}/ws`);

const scope = document.getElementById("scope");
const res = await fetch("/leaderboards");
const boards = await res.json();

scope.innerHTML += `<option value="Global">Global</option>`

for (const board of Object.keys(boards)) {
    scope.innerHTML += `<option value="${board}">${board}</option>`
}

scope.onchange = async () => {
    const res = await fetch("/leaderboard", {
        method: "PUT",
        body: JSON.stringify({
            active_leaderboard: scope.value
        })
    })

    console.log(res.status, res.statusText)
}

let activeLb = [];
let activeName = "Global";

ws.onopen = () => {
    console.log("Connected!");
};

ws.onmessage = async (event) => {
    const ev = JSON.parse(event.data);

    if (ev.type == "Beatmap") {
        await calcPerformance(ev.data);
        setBeatmap(ev.data);
        if (ev.data.mapID != activeLb[0]?.beatmap_id) activeLb = [];
    } else if (ev.type == "Leaderboard") {
        if (activeName !== ev.leaderboard) activeLb = [];
        activeLb.push(...ev.data);
        scope.value = ev.leaderboard;
        activeName = ev.leaderboard;
    }

    setLeaderboard(activeLb.sort((a, b) => b.total_score - a.total_score));
    console.log(ev);
};

ws.onclose = () => {
    console.log("Disconnected.");
};

ws.onerror = (err) => {
    console.error("WebSocket error:", err);
};


const PP_ACCS = [95, 97, 98, 99, 100];

async function calcPerformance(bmp) {
    const res = await fetch("/lazer/" + bmp.filename);
    const bytes = await res.text();

    let map = new Beatmap(bytes);

    if (map.isSuspicious()) {
        return
    }

    const diff = new Difficulty();
    const attr = diff.calculate(map);

    bmp.stars = attr.stars.toFixed(2);

    const difficultyAttrs = new Performance({ mods: 0 }).calculate(map);

    bmp.pp_stats = {};

    for (const acc of PP_ACCS) {
        const perf = new Performance({
            mods: 0,
            accuracy: acc,
            misses: 0,
            combo: difficultyAttrs.difficulty.maxCombo, // full combo
            hitresultPriority: HitResultPriority.BestCase,
        });

        const attrs = perf.calculate(difficultyAttrs);
        bmp.pp_stats[acc + "%"] = attrs.pp;

        perf.free(); 
        attrs.free();
    }

    difficultyAttrs.free();
    attr.free();
    map.free();
    diff.free();
}
