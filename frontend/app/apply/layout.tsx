import { ThemeProvider } from "@/lib/theme";

export default function ApplyLayout({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider>
      {children}
    </ThemeProvider>
  );
}
