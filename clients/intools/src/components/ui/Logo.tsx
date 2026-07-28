export function Logo({ collapsed = false }: { collapsed?: boolean }) {
  return (
    <div className="flex items-center gap-2">
      <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary font-display text-sm font-bold text-white">
        I
      </div>
      {!collapsed && (
        <span className="font-display text-lg font-bold text-white">
          SaleForges Intools
        </span>
      )}
    </div>
  );
}
