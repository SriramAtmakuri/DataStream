import { useMetrics } from '../hooks/useMetrics'
import { MetricCard } from './MetricCard'
import { OrdersChart } from './OrdersChart'
import { RevenueChart } from './RevenueChart'
import { TopProductsChart } from './TopProductsChart'

const PIPELINE_STEPS = ['Producer', 'Kafka', 'Processor', 'ClickHouse']

const SERVICES = [
  { label: 'Kafka UI',    port: 8090, path: '' },
  { label: 'Dagster',    port: 3000, path: '' },
  { label: 'Grafana',    port: 3001, path: '' },
  { label: 'Prometheus', port: 9090, path: '' },
  { label: 'Go Metrics', port: 8080, path: '/metrics' },
] as const

export function Dashboard() {
  const { ordersPerMinute, revenueByRegion, topProducts, errorRate, loading, lastUpdated } =
    useMetrics()

  const latestMinute = ordersPerMinute[ordersPerMinute.length - 1]
  const totalOrdersLastMin = latestMinute?.order_count ?? 0
  const totalRevenueLastMin = latestMinute?.total_revenue ?? 0
  const errorRateValue = errorRate?.error_rate ?? 0
  const totalRevenueRegions = revenueByRegion.reduce((s, r) => s + Number(r.revenue), 0)

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      {/* Header */}
      <header className="sticky top-0 z-10 border-b border-slate-800 bg-slate-900/95 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-600">
              <BoltIcon />
            </div>
            <div className="flex items-baseline gap-2">
              <span className="text-base font-semibold text-white">DataStream</span>
              <span className="hidden text-xs text-slate-500 sm:inline">E-Commerce Analytics</span>
            </div>
          </div>

          <div className="flex items-center gap-5">
            <div className="hidden items-center gap-1.5 text-xs text-slate-500 sm:flex">
              {PIPELINE_STEPS.map((step, i) => (
                <span key={step} className="flex items-center gap-1.5">
                  <span className="text-slate-300">{step}</span>
                  {i < PIPELINE_STEPS.length - 1 && <span className="text-slate-700">→</span>}
                </span>
              ))}
            </div>
            <div
              className={`flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium ${
                loading
                  ? 'bg-amber-500/10 text-amber-400'
                  : 'bg-emerald-500/10 text-emerald-400'
              }`}
            >
              <span
                className={`h-1.5 w-1.5 rounded-full ${
                  loading ? 'animate-pulse bg-amber-400' : 'bg-emerald-400'
                }`}
              />
              {loading ? 'Updating…' : `Live · ${lastUpdated?.toLocaleTimeString()}`}
            </div>
          </div>
        </div>
      </header>

      {/* Main */}
      <main className="mx-auto max-w-7xl px-6 py-8">
        {/* Section label */}
        <p className="mb-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
          Overview
        </p>

        {/* KPI Cards */}
        <div className="mb-8 grid grid-cols-2 gap-4 sm:grid-cols-4">
          <MetricCard
            title="Orders / min"
            value={totalOrdersLastMin.toLocaleString()}
            subtitle="last minute window"
            color="blue"
            icon={<BagIcon />}
          />
          <MetricCard
            title="Revenue / min"
            value={`$${Math.round(totalRevenueLastMin).toLocaleString()}`}
            subtitle="last minute window"
            color="green"
            icon={<CurrencyIcon />}
          />
          <MetricCard
            title="Error Rate"
            value={`${errorRateValue.toFixed(1)}%`}
            subtitle={`${errorRate?.failed ?? 0} failed / ${errorRate?.total ?? 0} total`}
            color={errorRateValue > 10 ? 'red' : errorRateValue > 5 ? 'yellow' : 'green'}
            icon={<AlertIcon />}
          />
          <MetricCard
            title="Regional Revenue"
            value={`$${Math.round(totalRevenueRegions).toLocaleString()}`}
            subtitle={`across ${revenueByRegion.length} region${revenueByRegion.length !== 1 ? 's' : ''}`}
            color="purple"
            icon={<GlobeIcon />}
          />
        </div>

        {/* Orders per Minute Chart */}
        <div className="mb-6 rounded-xl border border-slate-800 bg-slate-900 p-6">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-sm font-semibold text-slate-200">Orders per Minute</h2>
            <span className="text-xs text-slate-500">Last 60 minutes</span>
          </div>
          <OrdersChart data={ordersPerMinute} />
        </div>

        {/* Revenue + Top Products */}
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-200">Revenue by Region</h2>
              <span className="text-xs text-slate-500">Last hour</span>
            </div>
            <RevenueChart data={revenueByRegion} />
          </div>

          <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-200">Top Products</h2>
              <span className="text-xs text-slate-500">Last hour · by revenue</span>
            </div>
            <TopProductsChart data={topProducts} />
          </div>
        </div>

        {/* Services */}
        <div className="mt-8 rounded-xl border border-slate-800 bg-slate-900 p-5">
          <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-slate-500">
            Services
          </p>
          <div className="flex flex-wrap gap-2">
            {SERVICES.map(({ label, port, path }) => (
              <a
                key={label}
                href={`http://localhost:${port}${path}`}
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-1.5 rounded-lg border border-slate-700 bg-slate-800 px-3 py-1.5 text-xs text-slate-300 transition-colors hover:border-blue-500/50 hover:text-blue-400"
              >
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                {label}
                <span className="text-slate-600">:{port}</span>
              </a>
            ))}
          </div>
        </div>
      </main>
    </div>
  )
}

function BoltIcon() {
  return (
    <svg className="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z" />
    </svg>
  )
}

function BagIcon() {
  return (
    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 10.5V6a3.75 3.75 0 10-7.5 0v4.5m11.356-1.993l1.263 12c.07.665-.45 1.243-1.119 1.243H4.25a1.125 1.125 0 01-1.12-1.243l1.264-12A1.125 1.125 0 015.513 7.5h12.974c.576 0 1.059.435 1.119 1.007z" />
    </svg>
  )
}

function CurrencyIcon() {
  return (
    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  )
}

function AlertIcon() {
  return (
    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
    </svg>
  )
}

function GlobeIcon() {
  return (
    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={2} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" />
    </svg>
  )
}
