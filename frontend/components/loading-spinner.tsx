// components/ui/loading-spinner.tsx
import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";
import { Loader2 } from "lucide-react";

const spinnerVariants = cva(
  "animate-spin text-muted-foreground",
  {
    variants: {
      size: {
        default: "h-8 w-8",
        sm: "h-4 w-4",
        lg: "h-12 w-12",
        xl: "h-16 w-16"
      },
      variant: {
        default: "text-muted-foreground",
        primary: "text-primary",
        secondary: "text-secondary",
      }
    },
    defaultVariants: {
      size: "default",
      variant: "default"
    }
  }
);

export interface LoadingSpinnerProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof spinnerVariants> {
  text?: string;
  fullScreen?: boolean;
}

const LoadingSpinner = React.forwardRef<HTMLDivElement, LoadingSpinnerProps>(
  ({ className, size, variant, text, fullScreen, ...props }, ref) => {
    const spinnerContent = (
      <div
        ref={ref}
        className={cn(
          "flex flex-col items-center justify-center gap-2",
          fullScreen && "fixed inset-0 bg-background/80",
          className
        )}
        {...props}
      >
        <Loader2 className={cn(spinnerVariants({ size, variant }))} />
        {text && (
          <p className="text-sm text-muted-foreground animate-pulse">
            {text}
          </p>
        )}
      </div>
    );

    // If fullScreen is true, render in a portal to ensure it's always on top
    if (fullScreen) {
      return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          {spinnerContent}
        </div>
      );
    }

    return spinnerContent;
  }
);

LoadingSpinner.displayName = "LoadingSpinner";

export { LoadingSpinner, spinnerVariants };

// Usage examples:
// <LoadingSpinner /> - Default spinner
// <LoadingSpinner size="sm" /> - Small spinner
// <LoadingSpinner variant="primary" text="Loading..." /> - Primary color with text
// <LoadingSpinner fullScreen /> - Full screen overlay spinner