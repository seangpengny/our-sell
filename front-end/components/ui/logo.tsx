import Link from "next/link";

export function Logo({ href = "/sign-in", dark = false }: Readonly<{ href?: string; dark?: boolean }>) {
  return (
    <Link href={href} className={`brand ${dark ? "brand--dark" : ""}`} aria-label="Our Sell home">
      <span className="brand__mark"><span>O</span><span>S</span></span>
      <span className="brand__word">our <strong>sell</strong></span>
    </Link>
  );
}
