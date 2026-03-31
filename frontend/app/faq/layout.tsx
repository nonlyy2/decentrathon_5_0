import { ThemeProvider } from "@/lib/theme";

export default function FaqLayout({ children }: { children: React.ReactNode }) {
  return <ThemeProvider>{children}</ThemeProvider>;
}
