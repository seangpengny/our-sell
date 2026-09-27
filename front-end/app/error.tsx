"use client";

import { Button } from "@/components/ui/button";
import { StorefrontState } from "@/components/marketplace/storefront-state";

export default function ErrorPage({
  retry,
}: Readonly<{ error: Error & { digest?: string }; retry: () => void }>) {
  return (
    <StorefrontState
      title="Something didn’t load"
      description="Please try again. If the problem continues, return to the marketplace."
    >
      <Button onClick={() => retry()}>Try again</Button>
    </StorefrontState>
  );
}
