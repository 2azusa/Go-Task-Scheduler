import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { api } from "@/lib/api";
import { Job, JobLog, User, Script, JobType, HttpMethod } from "@/types/api"; // 导入新类型
import { toast } from "sonner";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Activity, ArrowLeft, CheckCircle2, XCircle, Edit, ScrollText } from "lucide-react"; // 导入新图标
import { format } from 'date-fns';

const JobDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const jobId = id ? parseInt(id, 10) : 0;

  const [job, setJob] = useState<Job | null>(null);
  const [logs, setLogs] = useState<JobLog[]>([]);
  const [notifiedUsers, setNotifiedUsers] = useState<User[]>([]);
  // --- [新增] 关联脚本 State ---
  const [associatedScripts, setAssociatedScripts] = useState<Script[]>([]);
  
  const [isLoadingJob, setIsLoadingJob] = useState(true);
  const [isLoadingLogs, setIsLoadingLogs] = useState(true);
  const [isLoadingUsers, setIsLoadingUsers] = useState(false);
  // --- [新增] 脚本加载 State ---
  const [isLoadingScripts, setIsLoadingScripts] = useState(false);

  const [logPage, setLogPage] = useState(1);
  const [logTotal, setLogTotal] = useState(0);
  const [logStatusFilter, setLogStatusFilter] = useState<string>("all");

  useEffect(() => {
    if (jobId) fetchJobDetails();
  }, [jobId]);
  
  useEffect(() => {
    if (jobId) fetchJobLogs();
  }, [jobId, logPage, logStatusFilter]);

  // 当 job 加载成功后，并行获取关联的用户和脚本
  useEffect(() => {
    const fetchAssociatedData = () => {
      // 获取用户
      if (job?.notifyTo && job.notifyTo.length > 0) {
        setIsLoadingUsers(true);
        const userPromises = job.notifyTo.map(userId => api.findUserById({ id: userId }));
        Promise.allSettled(userPromises)
          .then(results => {
            const successfulUsers = results
              .filter(r => r.status === 'fulfilled')
              .map(r => (r as PromiseFulfilledResult<User>).value);
            setNotifiedUsers(successfulUsers);
          })
          .finally(() => setIsLoadingUsers(false));
      }

      // --- [新增] 获取脚本 ---
      if (job?.scriptId && job.scriptId.length > 0) {
        setIsLoadingScripts(true);
        const scriptPromises = job.scriptId.map(scriptId => api.findScriptById({ id: scriptId }));
        Promise.allSettled(scriptPromises)
            .then(results => {
                const successfulScripts = results
                    .filter(r => r.status === 'fulfilled')
                    .map(r => (r as PromiseFulfilledResult<Script>).value);
                setAssociatedScripts(successfulScripts);
            })
            .finally(() => setIsLoadingScripts(false));
      }
    };
    fetchAssociatedData();
  }, [job]);

  const fetchJobDetails = async () => {
    setIsLoadingJob(true);
    try {
      const jobData = await api.findJobById({ id: jobId });
      setJob(jobData);
    } catch (error) {
      if (error instanceof Error) toast.error(`Failed to load job details: ${error.message}`);
    } finally {
      setIsLoadingJob(false);
    }
  };

  const fetchJobLogs = async () => {
    setIsLoadingLogs(true);
    try {
      const success = logStatusFilter !== 'all' ? logStatusFilter === 'true' : undefined;
      const logData = await api.getJobLogs({ jobId, page: logPage, pageSize: 10, success });
      setLogs(logData.list || []);
      setLogTotal(logData.total);
    } catch (error) {
      if (error instanceof Error) toast.error(`Failed to load job logs: ${error.message}`);
    } finally {
      setIsLoadingLogs(false);
    }
  };
  
  const handleFilterChange = (value: string) => {
    setLogPage(1);
    setLogStatusFilter(value);
  };

  const formatTimestamp = (ts: number) => ts ? format(new Date(ts * 1000), 'yyyy-MM-dd HH:mm:ss') : 'N/A';

  if (isLoadingJob) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  if (!job) {
    return (
      <div className="text-center py-10">
        <h2 className="text-xl text-muted-foreground">Job not found.</h2>
        <Button variant="outline" onClick={() => navigate("/jobs")} className="mt-4"><ArrowLeft className="mr-2 h-4 w-4" /> Back to Jobs</Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Button variant="outline" onClick={() => navigate("/jobs")} className="mb-4"><ArrowLeft className="mr-2 h-4 w-4" /> Back to Jobs List</Button>
      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
              <div><CardTitle className="text-2xl">{job.name}</CardTitle><CardDescription>{job.note || "No description for this job."}</CardDescription></div>
              <Button variant="outline" onClick={() => navigate(`/jobs/${id}/edit`)}><Edit className="mr-2 h-4 w-4" /> Edit</Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-x-6 gap-y-4 text-sm">
            <div className="space-y-1"><p className="text-muted-foreground">Schedule (Cron)</p><p className="font-mono">{job.spec}</p></div>
            <div className="space-y-1"><p className="text-muted-foreground">Job Type</p><p>{job.jobType === JobType.JobTypeCmd ? "Shell" : "HTTP"}</p></div>
            
            {/* --- [新增] 条件性显示 HTTP Method --- */}
            {job.jobType === JobType.JobTypeHttp && (
                <div className="space-y-1">
                    <p className="text-muted-foreground">HTTP Method</p>
                    <p className="font-semibold">{job.httpMethod === HttpMethod.HttpMethodGet ? 'GET' : 'POST'}</p>
                </div>
            )}

            <div className="space-y-1"><p className="text-muted-foreground">Assigned Node</p><p className="cursor-pointer hover:underline" onClick={() => navigate(`/nodes/${job.runOn}`)}>{job.hostname || <span className="text-muted-foreground">Not Assigned</span>}</p></div>
            <div className="space-y-1"><p className="text-muted-foreground">Timeout</p><p>{job.timeout > 0 ? `${job.timeout}s` : "None"}</p></div>
            <div className="space-y-1"><p className="text-muted-foreground">Retries</p><p>{job.retryTimes} times ({job.retryInterval}s interval)</p></div>
            <div className="space-y-1"><p className="text-muted-foreground">Last Updated</p><p>{formatTimestamp(job.updated)}</p></div>
            
            <div className="space-y-1 md:col-span-2">
              <p className="text-muted-foreground">Notify Users</p>
              {isLoadingUsers ? (<p className="text-sm text-muted-foreground">Loading...</p>) : 
               notifiedUsers.length > 0 ? (<div className="flex flex-wrap gap-2">{notifiedUsers.map((user) => (<Badge key={user.id} variant="secondary" className="cursor-pointer font-normal hover:bg-muted" onClick={() => toast.info(`User detail page for "${user.username}" is not yet implemented.`)}>{`ID: ${user.id} - ${user.username}`}</Badge>))}</div>) : 
               (<p className="text-sm text-muted-foreground">None</p>)}
            </div>

            {/* --- [新增] 显示关联脚本 --- */}
            <div className="col-span-full space-y-1">
                <p className="text-muted-foreground">Associated Scripts</p>
                {isLoadingScripts ? (<p className="text-sm text-muted-foreground">Loading...</p>) :
                 associatedScripts.length > 0 ? (<div className="flex flex-wrap gap-2">{associatedScripts.map(script => (<Badge key={script.id} variant="outline" className="cursor-pointer hover:bg-accent" onClick={() => navigate(`/scripts/${script.id}`)}><ScrollText className="h-3 w-3 mr-1.5"/>{script.name}</Badge>))}</div>) :
                 (<p className="text-sm text-muted-foreground">None</p>)}
            </div>

            <div className="col-span-full space-y-1"><p className="text-muted-foreground">Command</p><pre className="p-3 bg-muted rounded-md font-mono text-xs overflow-x-auto"><code>{job.command}</code></pre></div>
          </div>
        </CardContent>
      </Card>
      
      {/* 日志卡片部分保持不变 */}
      <Card>
        <CardHeader><div className="flex justify-between items-center"><CardTitle>Execution Logs</CardTitle><div className="w-[180px]"><Select value={logStatusFilter} onValueChange={handleFilterChange}><SelectTrigger><SelectValue placeholder="Filter by status..." /></SelectTrigger><SelectContent><SelectItem value="all">All Statuses</SelectItem><SelectItem value="true">Success</SelectItem><SelectItem value="false">Failed</SelectItem></SelectContent></Select></div></div></CardHeader>
        <CardContent>
           <div className="border rounded-lg"><Table><TableHeader><TableRow><TableHead>Time</TableHead><TableHead>Status</TableHead><TableHead>Node</TableHead><TableHead>Output</TableHead></TableRow></TableHeader><TableBody>
                {isLoadingLogs ? (<TableRow><TableCell colSpan={4} className="text-center py-8"><Activity className="h-6 w-6 animate-spin text-primary mx-auto" /></TableCell></TableRow>) : 
                 logs.length === 0 ? (<TableRow><TableCell colSpan={4} className="text-center py-8 text-muted-foreground">No logs found.</TableCell></TableRow>) : 
                 (logs.map((log) => (<TableRow key={log.id}><TableCell className="text-xs">{formatTimestamp(log.start_time)}</TableCell><TableCell>{log.success ? (<Badge variant="outline" className="text-success border-success"><CheckCircle2 className="mr-1 h-3 w-3" /> Success</Badge>) : (<Badge variant="destructive"><XCircle className="mr-1 h-3 w-3" /> Failed</Badge>)}</TableCell><TableCell>{log.hostname}</TableCell><TableCell><Dialog><DialogTrigger asChild><Button variant="outline" size="sm">View</Button></DialogTrigger><DialogContent className="max-w-3xl"><DialogHeader><DialogTitle>Log Output</DialogTitle></DialogHeader><pre className="mt-2 p-4 bg-muted rounded-md text-xs overflow-auto max-h-[60vh]"><code>{log.output || "No output."}</code></pre></DialogContent></Dialog></TableCell></TableRow>)))}
           </TableBody></Table></div>
           {logTotal > 10 && (<div className="flex items-center justify-between mt-4"><p className="text-sm text-muted-foreground">Page {logPage} of {Math.ceil(logTotal / 10)}</p><div className="flex gap-2"><Button variant="outline" size="sm" onClick={() => setLogPage(logPage - 1)} disabled={logPage === 1}>Previous</Button><Button variant="outline" size="sm" onClick={() => setLogPage(logPage + 1)} disabled={logPage * 10 >= logTotal}>Next</Button></div></div>)}
        </CardContent>
      </Card>
    </div>
  );
};

export default JobDetail;