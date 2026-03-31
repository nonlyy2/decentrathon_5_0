import { ThemeProvider } from "@/lib/theme";

export default function TMALayout({ children }: { children: React.ReactNode }) {
  return <ThemeProvider>{children}</ThemeProvider>;
}
