import { useEffect, useState } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";
import { api } from "@/lib/api";
import { SystemInfo, Node } from "@/types/api";
import { toast } from "sonner";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Activity, ArrowLeft, Cpu, HardDrive, MemoryStick, Server } from "lucide-react";

const NodeDetail = () => {
  const { uuid } = useParams<{ uuid: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const node: Node | undefined = location.state?.node;

  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (!uuid) {
      toast.error("Node UUID is missing.");
      navigate("/nodes");
      return;
    }

    api.getSystemInfo({ uuid })
      .then(setSystemInfo)
      .catch(error => {
        if (error instanceof Error) toast.error(`Failed to load system info: ${error.message}`);
      })
      .finally(() => setIsLoading(false));

  }, [uuid, navigate]);

  if (isLoading) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }
  
  if (!systemInfo) {
    return (
       <div className="container mx-auto p-6 text-center">
         <h2 className="text-xl text-muted-foreground">Could not load system information for this node.</h2>
         <p className="text-sm text-muted-foreground">The node might be offline or unreachable.</p>
         <Button variant="outline" onClick={() => navigate("/nodes")} className="mt-4"><ArrowLeft className="mr-2 h-4 w-4" /> Back to Nodes</Button>
       </div>
    );
  }

  const { os, cpu, ram, disk } = systemInfo;

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Button variant="outline" onClick={() => navigate("/nodes")} className="mb-4"><ArrowLeft className="mr-2 h-4 w-4" /> Back to Nodes List</Button>
      
      <Card>
          <CardHeader>
              <CardTitle className="text-2xl">{node?.hostname || 'Node Details'}</CardTitle>
              <CardDescription className="font-mono">{uuid}</CardDescription>
          </CardHeader>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
          <Card>
              <CardHeader><CardTitle className="flex items-center gap-2"><Server className="h-5 w-5"/> OS & Machine</CardTitle></CardHeader>
              <CardContent className="space-y-2 text-sm">
                  <div className="flex justify-between"><span>OS:</span><span className="font-semibold">{os?.goos || 'N/A'}</span></div>
                  <div className="flex justify-between"><span>Go Version:</span><span className="font-mono text-xs">{os?.goVersion || 'N/A'}</span></div>
                  <div className="flex justify-between"><span>Goroutines:</span><span className="font-semibold">{os?.numGoroutine || 'N/A'}</span></div>
              </CardContent>
          </Card>
          <Card>
              <CardHeader><CardTitle className="flex items-center gap-2"><Cpu className="h-5 w-5"/> CPU ({cpu?.cores || 0} Cores)</CardTitle></CardHeader>
              <CardContent className="space-y-2">
                  {(cpu?.cpus ?? []).length > 0 ? (
                    (cpu.cpus).map((coreUsage, index) => (
                        <div key={index} className="flex items-center gap-4 text-sm">
                            <span className="w-12">Core {index + 1}</span>
                            <Progress value={coreUsage} className="flex-1"/>
                            <span className="w-12 text-right font-mono">{coreUsage.toFixed(1)}%</span>
                        </div>
                    ))
                  ) : (
                    <p className="text-sm text-muted-foreground">CPU usage data is currently unavailable.</p>
                  )}
              </CardContent>
          </Card>
          <Card>
              <CardHeader><CardTitle className="flex items-center gap-2"><MemoryStick className="h-5 w-5"/> Memory (RAM)</CardTitle></CardHeader>
              <CardContent className="space-y-3">
                  <Progress value={ram?.usedPercent || 0} />
                  <div className="text-sm flex justify-between text-muted-foreground">
                      <span>Used: <span className="font-semibold text-primary">{(ram?.usedMb || 0).toFixed(0)} MB</span></span>
                      <span>Total: <span className="font-semibold text-primary">{(ram?.totalMB || 0).toFixed(0)} MB</span></span>
                  </div>
              </CardContent>
          </Card>
           <Card>
              <CardHeader><CardTitle className="flex items-center gap-2"><HardDrive className="h-5 w-5"/> Disk Space</CardTitle></CardHeader>
              <CardContent className="space-y-3">
                  <Progress value={disk?.usedPercent || 0} />
                   <div className="text-sm flex justify-between text-muted-foreground">
                      <span>Used: <span className="font-semibold text-primary">{(disk?.usedGb || 0).toFixed(1)} GB</span></span>
                      <span>Total: <span className="font-semibold text-primary">{(disk?.totalGb || 0).toFixed(1)} GB</span></span>
                  </div>
              </CardContent>
          </Card>
      </div>
    </div>
  );
};

export default NodeDetail;