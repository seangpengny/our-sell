import { Clock3, Flame, KeyRound, ShieldCheck, Sparkles } from "lucide-react";
import type { ProductFeature } from "@/lib/types";

const featureIcons = {
  key: KeyRound,
  flame: Flame,
  clock: Clock3,
  shield: ShieldCheck,
  sparkles: Sparkles,
};

export function FeatureIcon({
  icon,
}: Readonly<{ icon: ProductFeature["icon"] }>) {
  const Icon = featureIcons[icon];
  return <Icon size={16} aria-hidden="true" />;
}

export function ProductFeatures({
  features,
  className = "",
}: Readonly<{ features: ProductFeature[]; className?: string }>) {
  return (
    <ul className={`product-features ${className}`}>
      {features.map((feature) => (
        <li key={feature.label}>
          <FeatureIcon icon={feature.icon} />
          <span>{feature.label}</span>
        </li>
      ))}
    </ul>
  );
}
