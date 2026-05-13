import axios from 'axios'
import type { ErrorRate, OrdersPerMinute, RevenueByRegion, TopProduct } from '../types'

const api = axios.create({
  baseURL: '/api',
  timeout: 10_000,
})

export async function fetchOrdersPerMinute(): Promise<OrdersPerMinute[]> {
  const { data } = await api.get<OrdersPerMinute[]>('/metrics/orders-per-minute')
  return data ?? []
}

export async function fetchRevenueByRegion(): Promise<RevenueByRegion[]> {
  const { data } = await api.get<RevenueByRegion[]>('/metrics/revenue-by-region')
  return data ?? []
}

export async function fetchTopProducts(): Promise<TopProduct[]> {
  const { data } = await api.get<TopProduct[]>('/metrics/top-products')
  return data ?? []
}

export async function fetchErrorRate(): Promise<ErrorRate> {
  const { data } = await api.get<ErrorRate>('/metrics/error-rate')
  return data
}
