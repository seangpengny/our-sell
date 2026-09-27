import { StorefrontState } from "@/components/marketplace/storefront-state";

export default function NotFound() {
  return (
    <StorefrontState
      title="We couldn’t find that page"
      description="The link may have changed. Head back to the marketplace to continue exploring."
    />
  );
}
