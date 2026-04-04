type Props = {
  page: number;
  total: number;
  pageSize: number;
  onPage: (page: number) => void;
};

export default function Pagination({ page, total, pageSize, onPage }: Props) {
  if (total === 0) return null;

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="flex items-center justify-end gap-3 mt-4 text-sm">
      <button
        onClick={() => onPage(page - 1)}
        disabled={page <= 1}
        className="px-3 py-1 rounded border border-gray-200 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Trước
      </button>
      <span className="text-gray-600">
        {page} / {totalPages}
      </span>
      <button
        onClick={() => onPage(page + 1)}
        disabled={page >= totalPages}
        className="px-3 py-1 rounded border border-gray-200 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Sau
      </button>
    </div>
  );
}
