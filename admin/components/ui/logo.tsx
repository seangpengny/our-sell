import Link from "next/link";

export function Logo({ href = "/" }: Readonly<{ href?: string }>) {
  return (
    <Link className="brand" href={href} aria-label="Our Sell admin home">
      <span className="brand__mark"><span>o</span><span>s</span></span>
      <span className="brand__word">our <strong>sell</strong></span>
    </Link>
  );
}
