'use client';

const CATEGORIES = ['food', 'travel', 'utilities', 'entertainment', 'health', 'other'];

interface Props {
  category: string;
  onCategoryChange: (cat: string) => void;
}

export function FilterBar({ category, onCategoryChange }: Props) {
  return (
    <div className="flex items-center gap-3">
      <label htmlFor="filter-category" className="text-sm font-medium text-gray-700">
        Filter:
      </label>
      <select
        id="filter-category"
        value={category}
        onChange={(e) => onCategoryChange(e.target.value)}
        className="border rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
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
