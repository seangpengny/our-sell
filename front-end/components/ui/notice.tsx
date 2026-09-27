import { CircleAlert, CircleCheck, Info } from "lucide-react";

export function Notice({
  children,
  variant = "error",
}: Readonly<{
  children: React.ReactNode;
  variant?: "error" | "success" | "info";
}>) {
  const Icon =
    variant === "error" ? CircleAlert : variant === "info" ? Info : CircleCheck;
  return (
    <div
      className={`notice notice--${variant}`}
      role={variant === "error" ? "alert" : "status"}
    >
      <Icon size={18} aria-hidden="true" />
      <span>{children}</span>
    </div>
  );
}
