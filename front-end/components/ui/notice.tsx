import { CircleAlert, CircleCheck } from "lucide-react";

export function Notice({ children, variant = "error" }: Readonly<{ children: React.ReactNode; variant?: "error" | "success" | "info" }>) {
  return <div className={`notice notice--${variant}`} role={variant === "error" ? "alert" : "status"}>{variant === "error" ? <CircleAlert size={17} /> : <CircleCheck size={17} />}<span>{children}</span></div>;
}
