import Link from "next/link";
import type { ReactNode } from "react";

const nav = [
  { href: "/", label: "Inicio" },
  { href: "/tariffs", label: "Tarifas y Cuotas" },
  { href: "/ota-selectors", label: "Selectores OTA" },
  { href: "/anti-piracy", label: "Anti-Piratería" },
  { href: "/treasury", label: "Tesorería" },
];

export function AdminShell({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div>
            <p className="text-xs uppercase tracking-[0.2em] text-accent">RadarKm</p>
            <h1 className="text-xl font-semibold">RadarKm Admin</h1>
          </div>
          <nav className="flex flex-wrap gap-4 text-sm">
            {nav.map((item) => (
              <Link key={item.href} href={item.href} className="hover:text-accent">
                {item.label}
              </Link>
            ))}
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">{children}</main>
    </div>
  );
}
