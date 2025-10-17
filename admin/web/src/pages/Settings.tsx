import { useEffect, useState } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { Progress } from "@/components/ui/progress";

const Settings = () => {
  const [systemInfo, setSystemInfo] = useState<any>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadSystemInfo();
  }, []);

  const loadSystemInfo = async () => {
    setLoading(true);
    try {
      const response = await api.getSystemInfo();
      if (response.code === 200) {
        setSystemInfo(response.data);
      }
    } catch (error) {
      toast.error("加载系统信息失败");
    } finally {
      setLoading(false);
    }
  };

  if (loading || !systemInfo) {
    return (
      <MainLayout>
        <div className="text-center">加载中...</div>
      </MainLayout>
    );
  }

  const { os, disk, cpu, ram } = systemInfo;

  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">系统信息</h1>
          <p className="text-muted-foreground">查看服务器的系统信息</p>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Runtime</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between">
                <span>OS:</span>
                <span>{os.goos}</span>
              </div>
              <div className="flex justify-between">
                <span>CPU Cores:</span>
                <span>{os.numCpu}</span>
              </div>
              <div className="flex justify-between">
                <span>Compiler:</span>
                <span>{os.compiler}</span>
              </div>
              <div className="flex justify-between">
                <span>Go Version:</span>
                <span>{os.goVersion}</span>
              </div>
              <div className="flex justify-between">
                <span>Goroutines:</span>
                <span>{os.numGoroutine}</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Disk</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between">
                <span>Total (GB):</span>
                <span>{disk.totalGb}</span>
              </div>
              <div className="flex justify-between">
                <span>Used (GB):</span>
                <span>{disk.usedGb}</span>
              </div>
              <Progress value={disk.usedPercent} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>CPU</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between">
                <span>Physical Cores:</span>
                <span>{cpu.cores}</span>
              </div>
              {cpu.cpus.map((cpuUsage: number, index: number) => (
                <div key={index}>
                  <div className="flex justify-between">
                    <span>Core {index}:</span>
                    <span>{cpuUsage.toFixed(0)}%</span>
                  </div>
                  <Progress value={cpuUsage} />
                </div>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>RAM</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex justify-between">
                <span>Total (MB):</span>
                <span>{ram.totalMb}</span>
              </div>
              <div className="flex justify-between">
                <span>Used (MB):</span>
                <span>{ram.usedMb}</span>
              </div>
              <Progress value={ram.usedPercent} />
            </CardContent>
          </Card>
        </div>
      </div>
    </MainLayout>
  );
};

export default Settings;
