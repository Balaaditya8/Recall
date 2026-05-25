"use client";

import { useEffect, useState } from "react";

type Decision = {
  type: string;
  summary: string;
  owner: string;
  deadline: string;
  confidence: string;
  channel: string;
  timestamp: string;
  user: string;
  created_at: string;
};

const TYPE_CONFIG: Record<string, { label: string; color: string; dot: string }> = {
  task:     { label: "Task",     color: "bg-blue-500/10 text-blue-400 border-blue-500/20",   dot: "bg-blue-400" },
  decision: { label: "Decision", color: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20", dot: "bg-emerald-400" },
  deadline: { label: "Deadline", color: "bg-rose-500/10 text-rose-400 border-rose-500/20",   dot: "bg-rose-400" },
  none:     { label: "Other",    color: "bg-zinc-500/10 text-zinc-400 border-zinc-500/20",   dot: "bg-zinc-400" },
};

function timeAgo(dateStr: string) {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

function DecisionCard({ d }: { d: Decision }) {
  const cfg = TYPE_CONFIG[d.type] ?? TYPE_CONFIG.none;

  return (
    <div className="group relative border border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04] rounded-2xl p-5 transition-all duration-200">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-3 flex-wrap">
            <span className={`inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full border ${cfg.color}`}>
              <span className={`w-1.5 h-1.5 rounded-full ${cfg.dot}`} />
              {cfg.label}
            </span>
            {d.confidence === "high" && (
              <span className="text-xs text-zinc-500 font-mono">high confidence</span>
            )}
            {d.confidence === "low" && (
              <span className="inline-flex items-center gap-1 text-xs text-amber-400/80 font-mono">
                <span>⚠</span> needs review
              </span>
            )}
          </div>

          <p className="text-white/90 text-sm font-medium leading-snug mb-3">
            {d.summary}
          </p>

          <div className="flex items-center gap-4 flex-wrap">
            {d.owner && d.owner !== "null" && (
              <span className="flex items-center gap-1.5 text-xs text-zinc-500">
                <span className="text-zinc-600">owner</span>
                <span className="text-zinc-300">{d.owner}</span>
              </span>
            )}
            {d.deadline && d.deadline !== "null" && (
              <span className="flex items-center gap-1.5 text-xs text-zinc-500">
                <span className="text-zinc-600">due</span>
                <span className="text-zinc-300">{d.deadline}</span>
              </span>
            )}
            {d.channel && (
              <span className="flex items-center gap-1.5 text-xs text-zinc-500">
                <span className="text-zinc-600">#</span>
                <span className="text-zinc-400 font-mono">{d.channel}</span>
              </span>
            )}
          </div>
        </div>

        <div className="text-right shrink-0">
          <span className="text-xs text-zinc-600 font-mono">{timeAgo(d.created_at)}</span>
        </div>
      </div>
    </div>
  );
}

export default function Home() {
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<string>("all");

  useEffect(() => {
    fetch("http://localhost:8080/decisions")
      .then((r) => r.json())
      .then((data) => {
        setDecisions(Array.isArray(data) ? data : []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  const filtered = filter === "all" ? decisions : decisions.filter((d) => d.type === filter);

  const counts = {
    all: decisions.length,
    task: decisions.filter((d) => d.type === "task").length,
    decision: decisions.filter((d) => d.type === "decision").length,
    deadline: decisions.filter((d) => d.type === "deadline").length,
  };

  return (
    <main className="min-h-screen bg-[#0a0a0a] text-white">
      {/* Subtle grid background */}
      <div
        className="fixed inset-0 pointer-events-none"
        style={{
          backgroundImage:
            "linear-gradient(rgba(255,255,255,0.015) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.015) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
        }}
      />

      <div className="relative max-w-2xl mx-auto px-6 py-16">
        {/* Header */}
        <div className="mb-12">
          <div className="flex items-center gap-2 mb-2">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-xs text-zinc-500 font-mono tracking-widest uppercase">Live</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-white mb-1">
            Recall
          </h1>
          <p className="text-zinc-500 text-sm">
            Decisions and commitments from your conversations.
          </p>
        </div>

        {/* Filters */}
        <div className="flex gap-2 mb-8 flex-wrap">
          {(["all", "task", "decision", "deadline"] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`text-xs px-3 py-1.5 rounded-full border font-medium transition-all ${
                filter === f
                  ? "bg-white text-black border-white"
                  : "bg-transparent text-zinc-500 border-white/10 hover:border-white/20 hover:text-zinc-300"
              }`}
            >
              {f === "all" ? "All" : TYPE_CONFIG[f].label}
              <span className={`ml-1.5 ${filter === f ? "text-black/50" : "text-zinc-600"}`}>
                {counts[f]}
              </span>
            </button>
          ))}
        </div>

        {/* Content */}
        {loading ? (
          <div className="flex items-center gap-3 text-zinc-600 py-12">
            <div className="w-4 h-4 border border-zinc-700 border-t-zinc-400 rounded-full animate-spin" />
            <span className="text-sm font-mono">fetching decisions...</span>
          </div>
        ) : filtered.length === 0 ? (
          <div className="py-12 text-center">
            <p className="text-zinc-600 text-sm">No decisions captured yet.</p>
            <p className="text-zinc-700 text-xs mt-1">Send a message in Slack to get started.</p>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {filtered.map((d, i) => (
              <DecisionCard key={i} d={d} />
            ))}
          </div>
        )}

        {/* Footer */}
        {!loading && decisions.length > 0 && (
          <p className="text-center text-zinc-700 text-xs font-mono mt-10">
            {decisions.length} decision{decisions.length !== 1 ? "s" : ""} captured
          </p>
        )}
      </div>
    </main>
  );
}
