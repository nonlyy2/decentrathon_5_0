import { Badge } from "@/components/ui/badge";

const statusConfig: Record<string, { variant: "default" | "secondary" | "destructive" | "outline"; className: string }> = {
  pending: { variant: "secondary", className: "" },
  analyzed: { variant: "default", className: "bg-blue-500 hover:bg-blue-600" },
  shortlisted: { variant: "default", className: "bg-green-500 hover:bg-green-600" },
  rejected: { variant: "destructive", className: "" },
  waitlisted: { variant: "outline", className: "border-yellow-500 text-yellow-600" },
};

export default function StatusBadge({ status }: { status: string }) {
  const config = statusConfig[status] || statusConfig.pending;
  return (
    <Badge variant={config.variant} className={config.className}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </Badge>
  );
}
