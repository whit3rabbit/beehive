export default function TasksLoading() {
    return (
      <div className="flex h-[calc(100vh-4rem)] items-center justify-center">
        <div className="space-y-4">
          <div className="h-8 w-32 bg-muted animate-pulse rounded" />
          <div className="h-96 w-full max-w-4xl bg-muted animate-pulse rounded" />
        </div>
      </div>
    );
  }