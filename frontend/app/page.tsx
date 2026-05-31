"use client";

import { useEffect, useState } from "react";

type Decision = {
  id: number;
  type: string;
  summary: string;
  owner: string;
  deadline: string;
  confidence: string;
  channel: string;
  timestamp: string;
  user: string;
  created_at: string;
  status: string;
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

function isNullish(v: string) {
  return !v || v === "null" || v === "<nil>";
}

function PendingCard({ d, onAction }: { d: Decision; onAction: () => void }) {
  const [summary, setSummary] = useState(d.summary);
  const [owner, setOwner] = useState(isNullish(d.owner) ? "" : d.owner);
  const [deadline, setDeadline] = useState(isNullish(d.deadline) ? "" : d.deadline);
  const [type, setType] = useState(d.type);
  const [loading, setLoading] = useState(false);

  const confirm = async () => {
    setLoading(true);
    await fetch(`http://localhost:8080/decisions/${d.id}/confirm`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ summary, owner, deadline, type }),
    });
    setLoading(false);
    onAction();
  };

  const dismiss = async () => {
    setLoading(true);
    await fetch(`http://localhost:8080/decisions/${d.id}/dismiss`, { method: "POST" });
    setLoading(false);
    onAction();
  };

  return (
    <div className="border border-amber-500/20 bg-amber-500/5 rounded-2xl p-5">
      <div className="flex items-center gap-2 mb-4">
        <span className="text-amber-400 text-xs">⚠</span>
        <span className="text-xs text-amber-400 font-medium">Needs review</span>
        <span className="text-xs text-zinc-600 ml-auto font-mono">{timeAgo(d.created_at)}</span>
      </div>
      <p className="text-xs text-zinc-500 mb-1">Extracted:</p>
      <p className="text-zinc-400 text-sm mb-4 italic">"{d.summary}"</p>
      <div className="flex flex-col gap-2 mb-4">
        <div>
          <label className="text-xs text-zinc-600 mb-1 block">Type</label>
          <select value={type} onChange={(e) => setType(e.target.value)}
            className="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-white/20">
            <option value="task">Task</option>
            <option value="decision">Decision</option>
            <option value="deadline">Deadline</option>
          </select>
        </div>
        <div>
          <label className="text-xs text-zinc-600 mb-1 block">Summary</label>
          <input value={summary} onChange={(e) => setSummary(e.target.value)}
            className="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-white/20"
            placeholder="What was decided or committed to?" />
        </div>
        <div className="flex gap-2">
          <div className="flex-1">
            <label className="text-xs text-zinc-600 mb-1 block">Owner</label>
            <input value={owner} onChange={(e) => setOwner(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-white/20"
              placeholder="Who owns this?" />
          </div>
          <div className="flex-1">
            <label className="text-xs text-zinc-600 mb-1 block">Deadline</label>
            <input value={deadline} onChange={(e) => setDeadline(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-white/20"
              placeholder="When?" />
          </div>
        </div>
      </div>
      <div className="flex gap-2">
        <button onClick={confirm} disabled={loading || !summary}
          className="flex-1 bg-white text-black text-sm font-medium py-2 rounded-lg hover:bg-zinc-200 transition-colors disabled:opacity-40">
          {loading ? "Saving..." : "Confirm"}
        </button>
        <button onClick={dismiss} disabled={loading}
          className="px-4 bg-white/5 text-zinc-400 text-sm py-2 rounded-lg hover:bg-white/10 transition-colors border border-white/10">
          Dismiss
        </button>
      </div>
    </div>
  );
}

function DecisionCard({ d }: { d: Decision }) {
  const cfg = TYPE_CONFIG[d.type] ?? TYPE_CONFIG.none;
  return (
    <div className="group border border-white/[0.06] bg-white/[0.02] hover:bg-white/[0.04] rounded-2xl p-5 transition-all duration-200">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-3 flex-wrap">
            <span className={`inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full border ${cfg.color}`}>
              <span className={`w-1.5 h-1.5 rounded-full ${cfg.dot}`} />
              {cfg.label}
            </span>
          </div>
          <p className="text-white/90 text-sm font-medium leading-snug mb-3">{d.summary}</p>
          <div className="flex items-center gap-4 flex-wrap">
            {!isNullish(d.owner) && (
              <span className="flex items-center gap-1.5 text-xs text-zinc-500">
                <span className="text-zinc-600">owner</span>
                <span className="text-zinc-300">{d.owner}</span>
              </span>
            )}
            {!isNullish(d.deadline) && (
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
        <span className="text-xs text-zinc-600 font-mono shrink-0">{timeAgo(d.created_at)}</span>
      </div>
    </div>
  );
}

export default function Home() {
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [loading, setLoading] = useState(true);
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [channelFilter, setChannelFilter] = useState<string>("all");

  const load = () => {
    fetch("http://localhost:8080/decisions")
      .then((r) => r.json())
      .then((data) => { setDecisions(Array.isArray(data) ? data : []); setLoading(false); })
      .catch(() => setLoading(false));
  };

  useEffect(() => {
    load();
    const interval = setInterval(load, 5000);
    return () => clearInterval(interval);
  }, []);

  const pending = decisions.filter((d) => d.status === "pending");
  const confirmed = decisions.filter((d) => d.status === "confirmed");

  // unique channels from confirmed decisions
  const channels = [...new Set(confirmed.map((d) => d.channel).filter(Boolean))];

  const filtered = confirmed
    .filter((d) => typeFilter === "all" || d.type === typeFilter)
    .filter((d) => channelFilter === "all" || d.channel === channelFilter);

  const typeCounts = {
    all: confirmed.length,
    task: confirmed.filter((d) => d.type === "task").length,
    decision: confirmed.filter((d) => d.type === "decision").length,
    deadline: confirmed.filter((d) => d.type === "deadline").length,
  };

  return (
    <main className="min-h-screen bg-[#0a0a0a] text-white">
      <div className="fixed inset-0 pointer-events-none" style={{
        backgroundImage: "linear-gradient(rgba(255,255,255,0.015) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.015) 1px, transparent 1px)",
        backgroundSize: "48px 48px",
      }} />

      <div className="relative max-w-2xl mx-auto px-6 py-16">
        {/* Header */}
        <div className="mb-12">
          <div className="flex items-center gap-2 mb-2">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span className="text-xs text-zinc-500 font-mono tracking-widest uppercase">Live</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight text-white mb-1">Recall</h1>
          <p className="text-zinc-500 text-sm">Decisions and commitments from your conversations.</p>
        </div>

        {/* Pending */}
        {pending.length > 0 && (
          <div className="mb-10">
            <p className="text-xs text-amber-400/80 font-mono uppercase tracking-widest mb-3">
              {pending.length} pending review
            </p>
            <div className="flex flex-col gap-3">
              {pending.map((d) => <PendingCard key={d.id} d={d} onAction={load} />)}
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="flex flex-col gap-3 mb-8">
          {/* Type filters */}
          <div className="flex gap-2 flex-wrap">
            {(["all", "task", "decision", "deadline"] as const).map((f) => (
              <button key={f} onClick={() => setTypeFilter(f)}
                className={`text-xs px-3 py-1.5 rounded-full border font-medium transition-all ${
                  typeFilter === f
                    ? "bg-white text-black border-white"
                    : "bg-transparent text-zinc-500 border-white/10 hover:border-white/20 hover:text-zinc-300"
                }`}>
                {f === "all" ? "All" : TYPE_CONFIG[f].label}
                <span className={`ml-1.5 ${typeFilter === f ? "text-black/50" : "text-zinc-600"}`}>
                  {typeCounts[f]}
                </span>
              </button>
            ))}
          </div>

          {/* Channel filters — only show if more than one channel */}
          {channels.length > 1 && (
            <div className="flex gap-2 flex-wrap">
              <button onClick={() => setChannelFilter("all")}
                className={`text-xs px-3 py-1.5 rounded-full border font-medium transition-all ${
                  channelFilter === "all"
                    ? "bg-zinc-700 text-white border-zinc-600"
                    : "bg-transparent text-zinc-600 border-white/10 hover:border-white/20 hover:text-zinc-400"
                }`}>
                All channels
              </button>
              {channels.map((ch) => (
                <button key={ch} onClick={() => setChannelFilter(ch)}
                  className={`text-xs px-3 py-1.5 rounded-full border font-medium transition-all font-mono ${
                    channelFilter === ch
                      ? "bg-zinc-700 text-white border-zinc-600"
                      : "bg-transparent text-zinc-600 border-white/10 hover:border-white/20 hover:text-zinc-400"
                  }`}>
                  #{ch}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Decisions */}
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
            {filtered.map((d) => <DecisionCard key={d.id} d={d} />)}
          </div>
        )}

        {!loading && confirmed.length > 0 && (
          <p className="text-center text-zinc-700 text-xs font-mono mt-10">
            {confirmed.length} decision{confirmed.length !== 1 ? "s" : ""} captured
          </p>
        )}
      </div>
    </main>
  );
}
