import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'
import type { Product } from '@/types'

const colors = ['#FF3F6C', '#111827', '#FF8FAC', '#6B7280', '#EB2F5C', '#D1D5DB', '#C91F49', '#374151']

export function CategoryChart({ products }: { products: Product[] }) {
  const categoryData = Object.entries(
    products.reduce<Record<string, number>>((acc, p) => {
      acc[p.category] = (acc[p.category] ?? 0) + 1
      return acc
    }, {})
  )
    .map(([name, value]) => ({ name, value }))
    .sort((a, b) => b.value - a.value)

  return (
    <Card>
      <CardHeader>
        <p className="font-display text-base font-semibold text-ink-900">
          Category Distribution
        </p>
        <p className="text-xs text-ink-500">
          Share of recommended products by category
        </p>
      </CardHeader>

      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={categoryData}
                dataKey="value"
                nameKey="name"
                innerRadius={55}
                outerRadius={90}
                paddingAngle={2}
              >
                {categoryData.map((entry, i) => (
                  <Cell
                    key={entry.name}
                    fill={colors[i % colors.length]}
                    stroke="white"
                    strokeWidth={2}
                  />
                ))}
              </Pie>

              <Tooltip
                contentStyle={{
                  borderRadius: 12,
                  border: '1px solid #ECECEC',
                  fontSize: 12,
                }}
                formatter={(value, name) => [`${value} products`, name]}
              />
            </PieChart>
          </ResponsiveContainer>
        </div>

        <div className="mt-2 grid grid-cols-2 gap-x-4 gap-y-2">
          {categoryData.slice(0, 8).map((entry, i) => (
            <div
              key={entry.name}
              className="flex items-center gap-2 text-xs text-ink-700"
            >
              <span
                className="h-2 w-2 shrink-0 rounded-full"
                style={{
                  backgroundColor: colors[i % colors.length],
                }}
              />
              <span className="truncate">{entry.name}</span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}