'use client';

const CATEGORIES = ['food', 'travel', 'utilities', 'entertainment', 'health', 'other'];

interface Props {
  category: string;
  onCategoryChange: (cat: string) => void;
}

export function FilterBar({ category, onCategoryChange }: Props) {
  return (
    <div className="flex items-center gap-3">
      <label htmlFor="filter-category" className="text-sm font-semibold text-slate-700">
        Filter:
      </label>
      <select
        id="filter-category"
        value={category}
        onChange={(e) => onCategoryChange(e.target.value)}
        className="border border-slate-300 rounded-lg bg-slate-50 px-4 py-2 text-sm font-medium focus:bg-white focus:outline-none focus:ring-4 focus:ring-indigo-500/20 focus:border-indigo-500 transition-all duration-200 text-slate-800 shadow-sm"
      >
        <option value="">All categories</option>
        {CATEGORIES.map((c) => (
          <option key={c} value={c}>
            {c.charAt(0).toUpperCase() + c.slice(1)}
          </option>
        ))}
      </select>
    </div>
  );
}
