import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'
import type { Product } from '@/types'

const buckets = [
  { label: '< ₹1.5k', min: 0, max: 1500 },
  { label: '₹1.5k–3k', min: 1500, max: 3000 },
  { label: '₹3k–4.5k', min: 3000, max: 4500 },
  { label: '₹4.5k–6k', min: 4500, max: 6000 },
  { label: '> ₹6k', min: 6000, max: Infinity },
]

export function BudgetChart({ products }: { products: Product[] }) {
  const budgetData = buckets.map((b) => ({
    label: b.label,
    count: products.filter(
      (p) => p.price >= b.min && p.price < b.max
    ).length,
  }))

  return (
    <Card>
      <CardHeader>
        <p className="font-display text-base font-semibold text-ink-900">
          Budget Analysis
        </p>
        <p className="text-xs text-ink-500">
          Number of products across each price range
        </p>
      </CardHeader>

      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={budgetData}
              barCategoryGap="28%"
            >
              <CartesianGrid
                vertical={false}
                stroke="#ECECEC"
              />

              <XAxis
                dataKey="label"
                tick={{
                  fontSize: 11,
                  fill: '#6B7280',
                }}
                axisLine={{
                  stroke: '#ECECEC',
                }}
                tickLine={false}
              />

              <YAxis
                tick={{
                  fontSize: 11,
                  fill: '#6B7280',
                }}
                axisLine={false}
                tickLine={false}
              />

              <Tooltip
                cursor={{
                  fill: 'rgba(255,63,108,0.06)',
                }}
                contentStyle={{
                  borderRadius: 12,
                  border: '1px solid #ECECEC',
                  fontSize: 12,
                }}
                formatter={(value) => [`${value} products`, 'Count']}
              />

              <Bar
                dataKey="count"
                fill="#FF3F6C"
                radius={[8, 8, 0, 0]}
                maxBarSize={48}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}