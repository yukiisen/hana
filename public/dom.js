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

const dom = {
    shell: document.querySelector(".osu-shell"),
    cover: $("cover"),
    coverLink: $("cover-link"),
    status: $("status"),
    gamemode: $("gamemode"),
    mapset: $("mapset-link"),
    title: $("title-link"),
    artist: $("artist-link"),
    creator: $("creator"),
    creatorLink: $("creator-link"),
    difficulty: $("difficulty"),
    stars: $("stars"),
    ar: $("ar"),
    cs: $("cs"),
    od: $("od"),
    hp: $("hp"),
    ppGrid: $("pp-grid"),
    list: $("leaderboard"),
    previewing: $("previewing"),
    bg: $("background"),
};

const MODES = ["osu! standard", "osu!taiko", "osu!catch", "osu!mania"];

const rankedStatus = [
    "Graveyard", 
    "Pending",   
    "WIP",      
    "Ranked",    
    "Approved",  
    "Qualified", 
    "Loved", 
];

const PP_ACCS = ["95", "97", "98", "99", "100"];

function ppValue(v) {
    const n = typeof v === "number" ? v : v && v.pp;
    const num = Number(n);
    return n == null || Number.isNaN(num) ? "—" : Math.round(num);
}

function renderPpStats(stats) {
    const norm = {};
    if (stats && typeof stats === "object") {
        for (const [k, v] of Object.entries(stats)) {
            if (v == null) continue;
            norm[String(k).replace(/%$/, "").trim()] = v;
        }
    }

    const extra = Object.keys(norm).filter(
        (k) => !PP_ACCS.includes(k) && !Number.isNaN(parseFloat(k)),
    );
    const accs = PP_ACCS.filter((a) => norm[a] != null)
        .concat(extra)
        .sort((a, b) => parseFloat(a) - parseFloat(b));
    const empty = !accs.length;
    const list = empty ? PP_ACCS : accs;

    dom.ppGrid.innerHTML = list
        .map(
            (acc) => `<div class="pp-cell">
        <b>${esc(acc)}%</b>
        <strong>${esc(empty ? "—" : ppValue(norm[acc]))} <small>pp</small></strong>
      </div>`,
        )
        .join("");
}


export function setBeatmap(b) {
    if (!b) return;
    const setId = b.setID || b.setId || 0;
    const mapId = b.mapID || b.mapId || 0;
    const cover = `/lazer/${b.backgroundFilename}`
    const mapsetUrl = setId
        ? `https://osu.ppy.sh/beatmapsets/${setId}`
        : "https://osu.ppy.sh";
    const titleUrl = mapId ? `${mapsetUrl}#${mapId}` : mapsetUrl;
    const artist = b.artist || b.artistOriginal || "";
    const creator = b.creator || "";

    dom.cover.src = cover;
    dom.coverLink.href = mapsetUrl;
    dom.status.textContent = rankedStatus[b.rankedStatus - 1] || "—";
    dom.gamemode.textContent =
        typeof b.gamemode === "number" ? MODES[b.gamemode] : b.gamemode || "";
    dom.mapset.textContent = setId ? `MAPSET #${setId}` : "";
    dom.mapset.href = mapsetUrl;
    dom.mapset.style.display = setId ? "" : "none";
    dom.title.textContent = b.titleOriginal || b.title || "";
    dom.title.href = titleUrl;
    dom.artist.textContent = artist;
    dom.artist.href = `https://osu.ppy.sh/beatmapsets?q=${encodeURIComponent(artist)}`;
    dom.artist.style.display = artist ? "" : "none";
    dom.creator.textContent = creator;
    dom.creatorLink.href = `https://osu.ppy.sh/users/${encodeURIComponent(creator)}`;
    dom.difficulty.textContent = b.difficulty ? `[${b.difficulty}]` : "";
    dom.stars.textContent = `☆ ${b.stars || "0.00"}`;
    dom.ar.textContent = b.ar != null ? b.ar : "—";
    dom.cs.textContent = b.cs != null ? b.cs : "—";
    dom.od.textContent = b.od != null ? b.od : "—";
    dom.hp.textContent = b.hp != null ? b.hp : "—";
    dom.bg.style.backgroundImage = cover ? `url("${cover}")` : "";
    renderPpStats(b.pp_stats);
}

function timeAgo(date) {
  // Convert input to Date object if it's not already
  const past = date instanceof Date ? date : new Date(date);
  const now = new Date();
  
  // Calculate difference in seconds
  const seconds = Math.floor((now - past) / 1000);
  
  // Handle future dates
  if (seconds < 0) {
    return 'in the future';
  }
  
  // Time intervals in seconds
  const intervals = {
    year: 31536000,
    month: 2592000,
    week: 604800,
    day: 86400,
    hour: 3600,
    minute: 60,
    second: 1
  };
  
  // Find the largest interval that fits
  for (const [unit, secondsInUnit] of Object.entries(intervals)) {
    const count = Math.floor(seconds / secondsInUnit);
    if (count >= 1) {
      return count === 1 
        ? `1 ${unit} ago` 
        : `${count} ${unit}s ago`;
    }
  }
  
  return 'just now';
}

function row(s, index) {
    const name = s.user?.username || "";
    const userUrl = `https://osu.ppy.sh/users/${encodeURIComponent(name)}`;
    const stats = s.statistics || {};

    const mods = (s.mods || [])
        .map((m) => `<span class="mod">${esc(m.acronym)}</span>`)
        .join("");

    const hit = (key) => stats[key] ?? 0;
    const great = hit("great") + hit("geki");
    const ok = hit("ok") + hit("katu");
    const meh = hit("meh");
    const miss = hit("miss");

    const score = s.total_score ?? 0;
    const pp = s.pp == null ? "-" : Math.round(s.pp);

    return `<article class="score-row" data-index="${index}">
      <div class="rank">#${esc(index + 1)}</div>
      <span class="grade">${esc(s.rank ?? "-")}</span>
      <img class="avatar" src="${esc(s.user?.avatar_url || `https://a.ppy.sh/${s.user?.id ?? ""}`)}" alt="" loading="lazy" />
      <div class="player">
        <a class="link player-name" href="${userUrl}" target="_blank" rel="noreferrer"><strong>${esc(name)}</strong></a>
        <span>${esc(s.user?.country_code || "")}${s.ended_at ? ` <em>${timeAgo(s.ended_at)}</em>` : ""}</span>
      </div>
      <div class="row-stats">
        <div><b>MAX COMBO</b><strong class="combo">${esc(s.max_combo ?? 0)}x</strong></div>
        <div><b>ACCURACY</b><strong class="accuracy">${esc(((s.accuracy || 0) * 100).toFixed(2))}%</strong></div>
        <div class="hit-summary"><b>HITS</b><span><i class="great">${esc(great)}</i> <i class="ok">${esc(ok)}</i> <i class="meh">${esc(meh)}</i> <i class="miss">${esc(miss)}</i></span></div>
      </div>
      <div class="score-block">
        <a class="link score-link" href="${userUrl}/scores" target="_blank" rel="noreferrer">${esc(score.toLocaleString())}</a>
        ${mods ? `<div class="score-mods">${mods}</div>` : ""}
      </div>
      <div class="performance"><b>Performance</b><span class="pp">${esc(pp)} <small>pp</small></span></div>
    </article>`;
}

export function setLeaderboard(scores) {
    const list = Array.isArray(scores) ? scores : [];
    dom.list.innerHTML = list.length
        ? list.map(row).join("")
        : '<p class="empty">No scores to display.</p>';
}

dom.list.addEventListener("click", (event) => {
    const selected = event.target.closest(".score-row");
    if (!selected) return;
    dom.list
        .querySelectorAll(".score-row")
        .forEach((el) => el.classList.toggle("selected", el === selected));
    dom.previewing.textContent =
        selected.querySelector(".player-name").textContent;
});

setLeaderboard([]);
renderPpStats();
