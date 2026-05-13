import type { TopProduct } from '../types'

interface Props {
  data: TopProduct[]
}

export function TopProductsChart({ data }: Props) {
  if (data.length === 0) {
    return (
      <div className="flex h-48 items-center justify-center text-sm text-slate-500">
        Waiting for stream data…
      </div>
    )
  }

  const maxRevenue = Math.max(...data.map((d) => Number(d.revenue)))

  return (
    <div className="space-y-3">
      {data.map((item, idx) => {
        const revenue = Number(item.revenue)
        const pct = maxRevenue > 0 ? (revenue / maxRevenue) * 100 : 0
        return (
          <div key={item.product} className="flex items-center gap-3">
            <span className="w-5 shrink-0 text-right text-xs font-bold text-slate-500">
              {idx + 1}
            </span>
            <div className="min-w-0 flex-1">
              <div className="flex items-baseline justify-between gap-2">
                <span className="truncate text-sm font-medium text-slate-200">{item.product}</span>
                <span className="shrink-0 text-xs text-slate-400">${revenue.toLocaleString()}</span>
              </div>
              <div className="mt-1 h-1.5 w-full rounded-full bg-slate-800">
                <div
                  className="h-1.5 rounded-full bg-blue-500 transition-all duration-500"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <p className="mt-0.5 text-xs text-slate-600">
                {item.category} · {item.quantity} units
              </p>
            </div>
          </div>
        )
      })}
    </div>
  )
}
