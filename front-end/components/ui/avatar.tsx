import { initials } from "@/lib/format";

export function Avatar({ name, size = "md" }: Readonly<{ name: string; size?: "sm" | "md" | "lg" }>) {
  return <span className={`avatar avatar--${size}`} aria-hidden="true">{initials(name)}</span>;
}
