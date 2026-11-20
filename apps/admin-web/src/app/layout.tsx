import type { Metadata } from "next";
import { AdminShell } from "@/components/AdminShell";
import "./globals.css";

export const metadata: Metadata = {
  title: "RadarKm Admin",
  description: "Súper Administrador RadarKm — tarifas, OTA, antipiratería y tesorería",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="es">
      <body>
        <AdminShell>{children}</AdminShell>
      </body>
    </html>
  );
}
