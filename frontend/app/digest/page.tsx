"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

type Digest = {
  id: number;
  date: string;
  channel: string;
  summary: string;
  created_at: string;
};

type GroupedDigests = {
  [date: string]: Digest[];
};

function formatDate(dateStr: string) {
  const d = new Date(dateStr);
  const today = new Date();
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);

  if (d.toDateString() === today.toDateString()) return "Today";
  if (d.toDateString() === yesterday.toDateString()) return "Yesterday";
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

function formatSummary(summary: string) {
  return summary.split("\n").filter(Boolean).map((line, i) => {
    const isBullet = line.startsWith("-") || line.startsWith("•");
    return (
      <p key={i} className={`text-sm leading-relaxed ${isBullet ? "text-zinc-300 pl-2" : "text-zinc-400"}`}>
        {line}
      </p>
    );
  });
}

export default function DigestPage() {
  const [digests, setDigests] = useState<Digest[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [selectedDigest, setSelectedDigest] = useState<Digest | null>(null);

  useEffect(() => {
    fetch("http://localhost:8080/digests")
      .then((r) => r.json())
      .then((data) => {
        const arr = Array.isArray(data) ? data : [];
        setDigests(arr);
        if (arr.length > 0) {
          const firstDate = arr[0].date.split("T")[0];
          setSelectedDate(firstDate);
          setSelectedDigest(arr[0]);
        }
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  // group by date
  const grouped: GroupedDigests = digests.reduce((acc, d) => {
    const date = d.date.split("T")[0];
    if (!acc[date]) acc[date] = [];
    acc[date].push(d);
    return acc;
  }, {} as GroupedDigests);

  const dates = Object.keys(grouped).sort((a, b) => b.localeCompare(a));

  return (
    <main className="min-h-screen bg-[#0a0a0a] text-white flex">
      <div className="fixed inset-0 pointer-events-none" style={{
        backgroundImage: "linear-gradient(rgba(255,255,255,0.015) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.015) 1px, transparent 1px)",
        backgroundSize: "48px 48px",
      }} />

      {/* Sidebar */}
      <div className="relative w-64 shrink-0 border-r border-white/[0.06] h-screen sticky top-0 flex flex-col">
        <div className="p-6 border-b border-white/[0.06]">
          <Link href="/" className="text-xs text-zinc-600 hover:text-zinc-400 transition-colors font-mono mb-4 block">
            ← back to log
          </Link>
          <h2 className="text-sm font-semibold text-white">Daily Digests</h2>
          <p className="text-xs text-zinc-600 mt-0.5">AI summaries by day</p>
        </div>

        <div className="flex-1 overflow-y-auto p-3">
          {loading ? (
            <div className="flex items-center gap-2 p-3 text-zinc-600">
              <div className="w-3 h-3 border border-zinc-700 border-t-zinc-400 rounded-full animate-spin" />
              <span className="text-xs font-mono">loading...</span>
            </div>
          ) : dates.length === 0 ? (
            <p className="text-xs text-zinc-600 p-3">No digests yet.</p>
          ) : (
            dates.map((date) => (
              <div key={date} className="mb-4">
                <p className="text-xs text-zinc-600 font-mono px-3 mb-1">{formatDate(date)}</p>
                {grouped[date].map((d) => (
                  <button
                    key={d.id}
                    onClick={() => { setSelectedDate(date); setSelectedDigest(d); }}
                    className={`w-full text-left px-3 py-2.5 rounded-xl transition-all mb-1 ${
                      selectedDigest?.id === d.id
                        ? "bg-white/[0.06] text-white"
                        : "text-zinc-500 hover:bg-white/[0.03] hover:text-zinc-300"
                    }`}
                  >
                    <div className="flex items-center gap-2">
                      <span className="text-zinc-600 text-xs">#</span>
                      <span className="text-xs font-mono truncate">{d.channel}</span>
                    </div>
                  </button>
                ))}
              </div>
            ))
          )}
        </div>
      </div>

      {/* Content */}
      <div className="relative flex-1 p-12 max-w-2xl">
        {!selectedDigest ? (
          <div className="flex items-center justify-center h-full">
            <p className="text-zinc-600 text-sm">Select a digest to view</p>
          </div>
        ) : (
          <>
            <div className="mb-8">
              <p className="text-xs text-zinc-600 font-mono mb-1">
                {formatDate(selectedDigest.date.split("T")[0])} · #{selectedDigest.channel}
              </p>
              <h1 className="text-2xl font-bold text-white">Daily Digest</h1>
            </div>

            <div className="border border-white/[0.06] bg-white/[0.02] rounded-2xl p-6">
              <div className="flex flex-col gap-2">
                {formatSummary(selectedDigest.summary)}
              </div>
            </div>

            <p className="text-xs text-zinc-700 font-mono mt-6">
              generated {new Date(selectedDigest.created_at).toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" })}
            </p>
          </>
        )}
      </div>
    </main>
  );
}
