import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { MainLayout } from "@/components/layout/MainLayout";
import { api } from "@/lib/api";
import { SystemStatistics, WeekStatistics } from "@/types/api";
import { Server, CheckCircle, XCircle, Clock, AlertCircle } from "lucide-react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

const Dashboard = () => {
  const [stats, setStats] = useState<SystemStatistics | null>(null);
  const [weekStats, setWeekStats] = useState<WeekStatistics | null>(null);

  useEffect(() => {
    loadStatistics();
  }, []);

  const loadStatistics = async () => {
    try {
      const [todayRes, weekRes] = await Promise.all([
        api.getTodayStatistics(),
        api.getWeekStatistics(),
      ]);

      if (todayRes.code === 200) {
        setStats(todayRes.data);
      }
      if (weekRes.code === 200) {
        setWeekStats(weekRes.data);
      }
    } catch (error) {
      console.error("Failed to load statistics:", error);
    }
  };

  const chartData = (weekStats?.successDateCount || []).map((item, index) => ({
    date: item.date,
    成功: parseInt(item.count as unknown as string),
    失败: parseInt(weekStats?.failDateCount?.[index]?.count as unknown as string || "0"),
  }));

  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">仪表盘</h1>
          <p className="text-muted-foreground">系统运行状况概览</p>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                正常节点
              </CardTitle>
              <Server className="h-4 w-4 text-success" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-foreground">
                {stats?.normalNodeCount || 0}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                异常节点
              </CardTitle>
              <AlertCircle className="h-4 w-4 text-destructive" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-foreground">
                {stats?.failNodeCount || 0}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                成功执行
              </CardTitle>
              <CheckCircle className="h-4 w-4 text-success" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-foreground">
                {stats?.jobExcSuccessCount || 0}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                运行中
              </CardTitle>
              <Clock className="h-4 w-4 text-warning" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-foreground">
                {stats?.jobRunningCount || 0}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                执行失败
              </CardTitle>
              <XCircle className="h-4 w-4 text-destructive" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-foreground">
                {stats?.jobExcFailCount || 0}
              </div>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>最近7天任务执行趋势</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="成功"
                  stroke="hsl(var(--success))"
                  strokeWidth={2}
                />
                <Line
                  type="monotone"
                  dataKey="失败"
                  stroke="hsl(var(--destructive))"
                  strokeWidth={2}
                />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
};

export default Dashboard;