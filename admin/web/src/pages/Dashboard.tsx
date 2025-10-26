import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { TodayStatistics, WeekStatistics, SystemInfo } from "@/types/api";
import { Activity, CheckCircle2, XCircle, Server, PlayCircle, Calendar, FileText, ScrollText, Cpu, MemoryStick, HardDrive } from "lucide-react";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const Dashboard = () => {
  const [stats, setStats] = useState<TodayStatistics | null>(null);
  const [weekStats, setWeekStats] = useState<WeekStatistics | null>(null);
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    fetchStatistics();
    const interval = setInterval(fetchStatistics, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchStatistics = async () => {
    try {
      const [statsData, weekStatsData, systemInfoData] = await Promise.all([
        api.getTodayStatistics(),
        api.getWeekStatistics(),
        api.getSystemInfo({})
      ]);
      setStats(statsData);
      setWeekStats(weekStatsData);
      setSystemInfo(systemInfoData);
    } catch (error) {
      if (error instanceof Error) toast.error(error.message);
      else toast.error("Failed to load dashboard data");
    } finally {
      setIsLoading(false);
    }
  };

  const averageCpuUsage = useMemo(() => {
    if (!systemInfo?.cpu?.cpus?.length) return 0;
    const total = systemInfo.cpu.cpus.reduce((acc, usage) => acc + usage, 0);
    return total / systemInfo.cpu.cpus.length;
  }, [systemInfo]);

  const weeklyChartData = useMemo(() => {
    if (!weekStats) return [];
    const dataMap = new Map<string, { date: string; success: number; failed: number }>();
    weekStats.successDateCount.forEach(item => {
      const date = item.date.substring(5);
      if (!dataMap.has(date)) dataMap.set(date, { date, success: 0, failed: 0 });
      dataMap.get(date)!.success = parseInt(item.count, 10);
    });
    weekStats.failDateCount.forEach(item => {
      const date = item.date.substring(5);
      if (!dataMap.has(date)) dataMap.set(date, { date, success: 0, failed: 0 });
      dataMap.get(date)!.failed = parseInt(item.count, 10);
    });
    return Array.from(dataMap.values()).sort((a, b) => a.date.localeCompare(b.date));
  }, [weekStats]);

  if (isLoading) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
        <div className="mb-6">
          <h2 className="text-3xl font-bold mb-2">Dashboard</h2>
          <p className="text-muted-foreground">Monitor your distributed task scheduling system</p>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <Button variant="outline" className="justify-start text-left" onClick={() => navigate("/jobs")}><Calendar className="h-4 w-4 mr-2 shrink-0" /> View All Jobs</Button>
            <Button variant="outline" className="justify-start text-left" onClick={() => navigate("/nodes")}><Server className="h-4 w-4 mr-2 shrink-0" /> Manage Nodes</Button>
            <Button variant="outline" className="justify-start text-left" onClick={() => navigate("/logs")}><FileText className="h-4 w-4 mr-2 shrink-0" /> View Execution Logs</Button>
            <Button variant="outline" className="justify-start text-left" onClick={() => navigate("/scripts")}><ScrollText className="h-4 w-4 mr-2 shrink-0" /> Manage Scripts</Button>
        </div>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-5">
            <Card><CardHeader className="flex flex-row items-center justify-between pb-2"><CardTitle className="text-sm font-medium">Online Nodes</CardTitle><Server className="h-4 w-4 text-success" /></CardHeader><CardContent><div className="text-3xl font-bold">{stats?.normalNodeCount || 0}</div><p className="text-xs text-muted-foreground mt-1">Healthy nodes running</p></CardContent></Card>
            <Card><CardHeader className="flex flex-row items-center justify-between pb-2"><CardTitle className="text-sm font-medium">Offline Nodes</CardTitle><Server className="h-4 w-4 text-destructive" /></CardHeader><CardContent><div className="text-3xl font-bold">{stats?.failNodeCount || 0}</div><p className="text-xs text-muted-foreground mt-1">Nodes need attention</p></CardContent></Card>
            <Card><CardHeader className="flex flex-row items-center justify-between pb-2"><CardTitle className="text-sm font-medium">Successful Jobs</CardTitle><CheckCircle2 className="h-4 w-4 text-success" /></CardHeader><CardContent><div className="text-3xl font-bold">{stats?.jobExcSuccessCount || 0}</div><p className="text-xs text-muted-foreground mt-1">Completed today</p></CardContent></Card>
            <Card><CardHeader className="flex flex-row items-center justify-between pb-2"><CardTitle className="text-sm font-medium">Running Jobs</CardTitle><PlayCircle className="h-4 w-4 text-primary" /></CardHeader><CardContent><div className="text-3xl font-bold">{stats?.jobRunningCount || 0}</div><p className="text-xs text-muted-foreground mt-1">Currently executing</p></CardContent></Card>
            <Card><CardHeader className="flex flex-row items-center justify-between pb-2"><CardTitle className="text-sm font-medium">Failed Jobs</CardTitle><XCircle className="h-4 w-4 text-destructive" /></CardHeader><CardContent><div className="text-3xl font-bold">{stats?.jobExcFailCount || 0}</div><p className="text-xs text-muted-foreground mt-1">Require investigation</p></CardContent></Card>
        </div>
        
        <div className="grid gap-6 lg:grid-cols-5">
            <Card className="lg:col-span-3">
                <CardHeader><CardTitle>Weekly Job Executions</CardTitle></CardHeader>
                <CardContent className="h-[300px] w-full pl-0">
                <ResponsiveContainer>
                    <BarChart data={weeklyChartData} margin={{ top: 5, right: 30, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="date" fontSize={12} tickLine={false} axisLine={false} />
                        <YAxis fontSize={12} tickLine={false} axisLine={false} />
                        <Tooltip contentStyle={{ backgroundColor: 'hsl(var(--background))', borderColor: 'hsl(var(--border))' }}/>
                        <Legend wrapperStyle={{ fontSize: '14px' }} />
                        <Bar dataKey="success" fill="hsl(var(--success))" name="Success" stackId="a" radius={[4, 4, 0, 0]} />
                        <Bar dataKey="failed" fill="hsl(var(--destructive))" name="Failed" stackId="a" radius={[4, 4, 0, 0]} />
                    </BarChart>
                </ResponsiveContainer>
                </CardContent>
            </Card>

            <Card className="lg:col-span-2">
                <CardHeader><CardTitle>Management Server Status</CardTitle><CardDescription>Real-time metrics of the main server.</CardDescription></CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label className="flex items-center gap-2 text-sm font-medium"><Cpu className="h-4 w-4"/> CPU Usage ({systemInfo?.cpu?.cores || 0} Cores)</Label>
                        <Progress value={averageCpuUsage} />
                        <p className="text-xs text-muted-foreground text-right font-mono">{averageCpuUsage.toFixed(1)}%</p>
                    </div>
                    <div className="space-y-2">
                        <Label className="flex items-center gap-2 text-sm font-medium"><MemoryStick className="h-4 w-4"/> Memory (RAM)</Label>
                        <Progress value={systemInfo?.ram?.usedPercent || 0} />
                        <div className="text-xs text-muted-foreground flex justify-between">
                        <span>Used: <span className="font-semibold text-primary">{(systemInfo?.ram?.usedMb || 0).toFixed(0)} MB</span></span>
                        <span>Total: <span className="font-semibold text-primary">{(systemInfo?.ram?.totalMB || 0).toFixed(0)} MB</span></span>
                        </div>
                    </div>
                    <div className="space-y-2">
                        <Label className="flex items-center gap-2 text-sm font-medium"><HardDrive className="h-4 w-4"/> Disk Space</Label>
                        <Progress value={systemInfo?.disk?.usedPercent || 0} />
                        <div className="text-xs text-muted-foreground flex justify-between">
                        <span>Used: <span className="font-semibold text-primary">{(systemInfo?.disk?.usedGb || 0).toFixed(1)} GB</span></span>
                        <span>Total: <span className="font-semibold text-primary">{(systemInfo?.disk?.totalGb || 0).toFixed(1)} GB</span></span>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
      </div>
  );
};

export default Dashboard;