(function () {
	"use strict";

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
		objects: $("objects"),
		plays: $("plays"),
		length: $("length"),
		list: $("leaderboard"),
		previewing: $("previewing"),
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

	function setBeatmap(b) {
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
		dom.objects.textContent = `${b.objectCount != null ? b.objectCount : "—"} objects`;
		dom.plays.textContent = `${b.plays != null ? b.plays : "—"} plays`;
		dom.length.textContent = `Audio ${b.length || "—:—"}`;
		dom.shell.style.backgroundImage = cover ? `url("${cover}")` : "";
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
            <span>${esc(s.user?.country_code || "")}${s.ended_at ? ` <em>${esc(s.ended_at)}</em>` : ""}</span>
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
          <div class="performance"><b>PERFORMANCE</b><span class="pp">${esc(pp)} <small>pp</small></span></div>
        </article>`;
    }

	function setLeaderboard(scores) {
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

	window.setBeatmap = setBeatmap;
	window.setLeaderboard = setLeaderboard;
})();

const ws = new WebSocket(`ws://${location.host}/ws`);

let activeLb = []

ws.onopen = () => {
    console.log("Connected!");
};

ws.onmessage = (event) => {
    const ev = JSON.parse(event.data);

    if (ev.type == "Beatmap") {
        window.setBeatmap(ev.data);
        activeLb = [];
    } else if (ev.type == "Leaderboard") {
        activeLb.push(...ev.data);
    }

    window.setLeaderboard(activeLb.sort((a, b) => b.total_score - a.total_score));
    console.log(ev);
};

ws.onclose = () => {
    console.log("Disconnected.");
};

ws.onerror = (err) => {
    console.error("WebSocket error:", err);
};
