interface MetricCardProps {
  title: string
  value: string | number
  subtitle?: string
  color?: 'blue' | 'green' | 'red' | 'yellow' | 'purple'
  icon?: React.ReactNode
}

const colorConfig = {
  blue:   { bg: 'bg-blue-500/10',    border: 'border-blue-500/20',    value: 'text-blue-400',    icon: 'bg-blue-500/20 text-blue-400' },
  green:  { bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', value: 'text-emerald-400', icon: 'bg-emerald-500/20 text-emerald-400' },
  red:    { bg: 'bg-red-500/10',     border: 'border-red-500/20',     value: 'text-red-400',     icon: 'bg-red-500/20 text-red-400' },
  yellow: { bg: 'bg-amber-500/10',   border: 'border-amber-500/20',   value: 'text-amber-400',   icon: 'bg-amber-500/20 text-amber-400' },
  purple: { bg: 'bg-violet-500/10',  border: 'border-violet-500/20',  value: 'text-violet-400',  icon: 'bg-violet-500/20 text-violet-400' },
}

export function MetricCard({ title, value, subtitle, color = 'blue', icon }: MetricCardProps) {
  const c = colorConfig[color]
  return (
    <div className={`rounded-xl border p-5 ${c.bg} ${c.border}`}>
      <div className="flex items-start justify-between">
        <p className="text-xs font-medium text-slate-400">{title}</p>
        {icon && (
          <div className={`flex h-7 w-7 items-center justify-center rounded-lg ${c.icon}`}>
            {icon}
          </div>
        )}
      </div>
      <p className={`mt-2 text-2xl font-bold ${c.value}`}>{value}</p>
      {subtitle && <p className="mt-1 text-xs text-slate-500">{subtitle}</p>}
    </div>
  )
}
