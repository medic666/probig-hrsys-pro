import { Timeline, Typography, Tag, Empty } from 'antd'
import dayjs from 'dayjs'

const { Text } = Typography

const eventTypeColors: Record<string, string> = {
  create: 'green', update: 'blue', delete: 'red',
  grant: 'purple', calculate: 'orange', carry_over: 'cyan', close_month: 'gold',
}

interface EventItem {
  id: number
  eventType: string
  createdAt: string
  remark: string
  operatorId: number
}

export default function EventTimeline({ events }: { events: EventItem[] }) {
  if (!events || events.length === 0) return <Empty description="暂无事件记录" />

  return (
    <Timeline
      items={events.map((e) => ({
        color: eventTypeColors[e.eventType] || 'gray',
        children: (
          <div>
            <Tag color={eventTypeColors[e.eventType] || 'default'}>{e.eventType}</Tag>
            <Text style={{ marginLeft: 8 }}>{e.remark}</Text>
            <br />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {dayjs(e.createdAt).format('YYYY-MM-DD HH:mm:ss')} | 操作人ID: {e.operatorId}
            </Text>
          </div>
        ),
      }))}
    />
  )
}
