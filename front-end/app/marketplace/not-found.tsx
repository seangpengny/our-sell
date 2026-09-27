import { StorefrontState } from "@/components/marketplace/storefront-state";

export default function ListingNotFound() {
  return (
    <StorefrontState
      title="This listing isn’t available"
      description="It may have been removed or is no longer for sale. Explore the marketplace for available products."
    />
  );
}
