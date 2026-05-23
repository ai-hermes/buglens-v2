import { Alert, Button, Card, Space, Typography } from 'antd'
import { simulateRumErrors } from '@/monitoring/rum'
import { useCounterStore } from '@/store/modules/counter'

export function HomePage() {
  const { count, increment, reset } = useCounterStore()

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      <Card title="React + Vite 全家桶模板">
        <Typography.Paragraph>
          已完成基础初始化：路由、Zustand 状态管理、Axios 请求层、Ant Design、工程化规范。
        </Typography.Paragraph>
        <Space>
          <Button type="primary" onClick={increment}>
            Count +1
          </Button>
          <Button onClick={reset}>Reset</Button>
          <Typography.Text>count: {count}</Typography.Text>
        </Space>
      </Card>

      <Card title="ARMS 异常模拟面板（测试专用）">
        <Typography.Paragraph type="secondary">
          点击按钮可快速模拟前端异常，便于在 ARMS 控制台验证采集链路。请勿在生产环境频繁触发。
        </Typography.Paragraph>
        <Space wrap>
          <Button danger onClick={simulateRumErrors.syncError}>
            触发 JS 同步异常
          </Button>
          <Button danger onClick={simulateRumErrors.unhandledRejection}>
            触发 Promise 未处理拒绝
          </Button>
          <Button danger onClick={simulateRumErrors.consoleError}>
            触发 Console 错误
          </Button>
          <Button danger onClick={simulateRumErrors.apiFailure}>
            触发 API 失败
          </Button>
          <Button danger onClick={simulateRumErrors.resourceFailure}>
            触发资源加载失败
          </Button>
        </Space>
      </Card>

      <Alert type="success" showIcon title="你现在可以直接在此基座上开发业务页面。" />
    </Space>
  )
}
