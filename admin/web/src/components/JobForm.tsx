import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm, Controller } from "react-hook-form";
import { api } from "@/lib/api";
import { JobUpdate, JobType, Allocation, NotifyType, User, Node, Script, HttpMethod } from "@/types/api"; // 导入新类型
import { toast } from "sonner";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { Badge } from "@/components/ui/badge";
import { Activity, ArrowLeft, ChevronsUpDown, Check, X } from "lucide-react";

// 定义一个更精确的表单数据类型
type JobFormData = Omit<JobUpdate, 'jobType' | 'allocation' | 'notifyType' | 'httpMethod'> & {
  jobType: string;
  allocation: string;
  notifyType: string;
  httpMethod: string;
};

const JobForm = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isEditMode = !!id;

  // --- 统一管理所有下拉列表的数据 ---
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [allNodes, setAllNodes] = useState<Node[]>([]);
  const [allScripts, setAllScripts] = useState<Script[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const { register, handleSubmit, control, reset, watch, setValue, formState: { errors } } = useForm<JobFormData>({
    defaultValues: {
      notifyTo: [],
      scriptId: [],
      jobType: String(JobType.JobTypeCmd),
      allocation: String(Allocation.ManualAllocation),
      notifyType: String(NotifyType.NotifyTypeMail),
      timeout: 0,
      retryTimes: 0,
      retryInterval: 10,
    }
  });

  const watchedJobType = watch('jobType');
  const watchedAllocation = watch('allocation');

  // 当 jobType 变化时，自动调整 allocation
  useEffect(() => {
    if (watchedJobType === String(JobType.JobTypeCmd)) {
      setValue('allocation', String(Allocation.ManualAllocation));
    }
  }, [watchedJobType, setValue]);

  // 并行获取所有初始化数据
  useEffect(() => {
    Promise.all([
      api.searchUsers({ pageSize: 1000 }),
      api.searchNodes({ pageSize: 1000 }),
      api.searchScripts({ pageSize: 1000 }),
      isEditMode ? api.findJobById({ id: parseInt(id!, 10) }) : Promise.resolve(null)
    ])
    .then(([usersResult, nodesResult, scriptsResult, jobData]) => {
      setAllUsers(usersResult.list || []);
      setAllNodes(nodesResult.list || []);
      setAllScripts(scriptsResult.list || []);
      
      if (isEditMode && jobData) {
        reset({
          ...jobData,
          jobType: String(jobData.jobType),
          allocation: String(jobData.allocation),
          notifyType: String(jobData.notifyType),
          httpMethod: String(jobData.httpMethod),
          notifyTo: jobData.notifyTo || [],
          scriptId: jobData.scriptId || [],
        });
      }
    })
    .catch(error => {
      toast.error(`Failed to load initial data: ${error.message}`);
      if (isEditMode) navigate("/jobs");
    })
    .finally(() => setIsLoading(false));
  }, [id, isEditMode, navigate, reset]);

  const onSubmit = async (data: JobFormData) => {
    const payload: JobUpdate = {
      ...data,
      id: isEditMode ? parseInt(id!, 10) : undefined,
      jobType: Number(data.jobType) as JobType,
      allocation: Number(data.allocation) as Allocation,
      notifyType: Number(data.notifyType) as NotifyType,
      timeout: Number(data.timeout) || 0,
      retryTimes: Number(data.retryTimes) || 0,
      retryInterval: Number(data.retryInterval) || 0,
      notifyTo: data.notifyTo,
      scriptId: data.scriptId,
      // 仅当类型为 HTTP 时才提交 httpMethod
      httpMethod: watchedJobType === String(JobType.JobTypeHttp) ? Number(data.httpMethod) as HttpMethod : undefined,
    };

    try {
      await api.saveJob(payload);
      toast.success(isEditMode ? "Job updated!" : "Job created!");
      navigate(isEditMode ? `/jobs/${id}` : "/jobs");
    } catch (error) {
      if (error instanceof Error) toast.error(error.message);
    }
  };

  if (isLoading) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Button variant="outline" onClick={() => navigate(isEditMode ? `/jobs/${id}` : "/jobs")} className="mb-4"><ArrowLeft className="mr-2 h-4 w-4" /> Back</Button>
      <Card>
        <CardHeader><CardTitle>{isEditMode ? "Edit Job" : "Create New Job"}</CardTitle><CardDescription>Fill out the form to configure the job.</CardDescription></CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            {/* --- 主配置 --- */}
            <div className="grid md:grid-cols-2 gap-6">
              <div><Label htmlFor="name">Job Name</Label><Input id="name" {...register("name", { required: "Name is required" })} />{errors.name && <p className="text-destructive text-sm mt-1">{errors.name.message}</p>}</div>
              <div><Label htmlFor="spec">Schedule (Cron)</Label><Input id="spec" {...register("spec", { required: "Cron spec is required" })} placeholder="* * * * *" />{errors.spec && <p className="text-destructive text-sm mt-1">{errors.spec.message}</p>}</div>
              <div><Label>Job Type</Label><Controller name="jobType" control={control} render={({ field }) => (<Select onValueChange={field.onChange} value={field.value}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value={String(JobType.JobTypeCmd)}>Shell</SelectItem><SelectItem value={String(JobType.JobTypeHttp)}>HTTP</SelectItem></SelectContent></Select>)} /></div>
              
              {/* --- 条件性 HTTP Method --- */}
              {watchedJobType === String(JobType.JobTypeHttp) && (
                <div>
                  <Label>HTTP Method</Label>
                  <Controller name="httpMethod" control={control} rules={{ required: "HTTP Method is required" }} render={({ field }) => (<Select onValueChange={field.onChange} value={field.value}><SelectTrigger><SelectValue placeholder="Select a method..." /></SelectTrigger><SelectContent><SelectItem value={String(HttpMethod.HttpMethodGet)}>GET</SelectItem><SelectItem value={String(HttpMethod.HttpMethodPost)}>POST</SelectItem></SelectContent></Select>)} />
                  {errors.httpMethod && <p className="text-destructive text-sm mt-1">{errors.httpMethod.message}</p>}
                </div>
              )}
              
              <div className="col-span-full"><Label htmlFor="command">Command / URL</Label><Textarea id="command" {...register("command", { required: "Command or URL is required" })} rows={4} />{errors.command && <p className="text-destructive text-sm mt-1">{errors.command.message}</p>}</div>
            </div>
            
            {/* --- 分配与执行 --- */}
            <h3 className="text-lg font-semibold pt-4 border-t">Allocation & Execution</h3>
            <div className="grid md:grid-cols-2 gap-6">
              <div><Label>Allocation</Label><Controller name="allocation" control={control} render={({ field }) => (<Select onValueChange={field.onChange} value={field.value} disabled={watchedJobType === String(JobType.JobTypeCmd)}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value={String(Allocation.ManualAllocation)}>Manual</SelectItem><SelectItem value={String(Allocation.AutoAllocation)}>Auto</SelectItem></SelectContent></Select>)} /></div>
              
              {/* --- [新增] Node Combobox --- */}
              <div>
                <Label htmlFor="runOn">Run On Node</Label>
                <Controller name="runOn" control={control} rules={{ required: watchedAllocation === String(Allocation.ManualAllocation) ? "Node is required for Manual allocation" : false }} render={({ field }) => (
                  <Popover><PopoverTrigger asChild><Button variant="outline" role="combobox" className="w-full justify-between">{allNodes.find(node => node.uuid === field.value)?.hostname || "Select a node..."}<ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" /></Button></PopoverTrigger>
                    <PopoverContent className="w-[--radix-popover-trigger-width] p-0"><Command><CommandInput placeholder="Search node..." /><CommandList><CommandEmpty>No node found.</CommandEmpty><CommandGroup>
                      {allNodes.map(node => (<CommandItem key={node.uuid} value={node.hostname} onSelect={() => field.onChange(node.uuid)}><Check className={`mr-2 h-4 w-4 ${field.value === node.uuid ? "opacity-100" : "opacity-0"}`} />{node.hostname} ({node.ip})</CommandItem>))}
                    </CommandGroup></CommandList></Command></PopoverContent>
                  </Popover>
                )} />
                {errors.runOn && <p className="text-destructive text-sm mt-1">{errors.runOn.message}</p>}
              </div>

              {/* --- [新增] 关联脚本 --- */}
              <div className="col-span-full"><Label>Associated Scripts</Label><Controller control={control} name="scriptId" render={({ field }) => (
                  <Popover><PopoverTrigger asChild><Button variant="outline" role="combobox" className="w-full justify-between h-auto min-h-10 flex-wrap"><div className="flex gap-1 flex-wrap">{field.value?.length > 0 ? allScripts.filter(s => field.value.includes(s.id)).map(s => (<Badge key={s.id} variant="secondary">{s.name}<button className="ml-1" onMouseDown={(e) => e.preventDefault()} onClick={() => field.onChange(field.value.filter(id => id !== s.id))}><X className="h-3 w-3" /></button></Badge>)) : <span>Select scripts...</span>}</div><ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50" /></Button></PopoverTrigger>
                    <PopoverContent className="w-[--radix-popover-trigger-width] p-0"><Command><CommandInput placeholder="Search script..." /><CommandList><CommandEmpty>No script found.</CommandEmpty><CommandGroup>
                      {allScripts.map(script => (<CommandItem key={script.id} value={script.name} onSelect={() => field.onChange(field.value.includes(script.id) ? field.value.filter(id => id !== script.id) : [...field.value, script.id])}><Check className={`mr-2 h-4 w-4 ${field.value.includes(script.id) ? "opacity-100" : "opacity-0"}`} />{script.name}</CommandItem>))}
                    </CommandGroup></CommandList></Command></PopoverContent>
                  </Popover>
              )} /></div>
            </div>

            {/* --- 超时与重试 --- */}
            <h3 className="text-lg font-semibold pt-4 border-t">Timeout & Retries</h3>
            <div className="grid md:grid-cols-3 gap-6">
              <div><Label htmlFor="timeout">Timeout (seconds)</Label><Input id="timeout" type="number" {...register("timeout")} placeholder="0 for no timeout"/></div>
              <div><Label htmlFor="retryTimes">Retry Times</Label><Input id="retryTimes" type="number" {...register("retryTimes")} /></div>
              <div><Label htmlFor="retryInterval">Retry Interval (seconds)</Label><Input id="retryInterval" type="number" {...register("retryInterval")} /></div>
            </div>

            {/* --- 通知 --- */}
            <h3 className="text-lg font-semibold pt-4 border-t">Notifications</h3>
            <div className="grid md:grid-cols-2 gap-6">
              <div><Label>Notify Type</Label><Controller name="notifyType" control={control} render={({ field }) => (<Select onValueChange={field.onChange} value={field.value}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value={String(NotifyType.NotifyTypeMail)}>Email</SelectItem><SelectItem value={String(NotifyType.NotifyTypeWebHook)}>Webhook</SelectItem></SelectContent></Select>)} /></div>
              <div className="col-span-full"><Label>Notify Users</Label><Controller control={control} name="notifyTo" render={({ field }) => (
                <Popover><PopoverTrigger asChild><Button variant="outline" role="combobox" className="w-full justify-between h-auto min-h-10 flex-wrap"><div className="flex gap-1 flex-wrap">{field.value?.length > 0 ? allUsers.filter(u => field.value.includes(u.id)).map(u => (<Badge key={u.id} variant="secondary">{u.username}<button className="ml-1" onMouseDown={(e) => e.preventDefault()} onClick={() => field.onChange(field.value.filter(id => id !== u.id))}><X className="h-3 w-3" /></button></Badge>)) : <span>Select users...</span>}</div><ChevronsUpDown className="h-4 w-4 shrink-0 opacity-50" /></Button></PopoverTrigger>
                  <PopoverContent className="w-[--radix-popover-trigger-width] p-0"><Command><CommandInput placeholder="Search users..." /><CommandList><CommandEmpty>No users found.</CommandEmpty><CommandGroup>
                    {allUsers.map(user => (<CommandItem key={user.id} value={user.username} onSelect={() => field.onChange(field.value.includes(user.id) ? field.value.filter(id => id !== user.id) : [...field.value, user.id])}><Check className={`mr-2 h-4 w-4 ${field.value.includes(user.id) ? "opacity-100" : "opacity-0"}`} />{user.username}</CommandItem>))}
                  </CommandGroup></CommandList></Command></PopoverContent>
                </Popover>
              )} /></div>
              <div className="col-span-full"><Label htmlFor="note">Notes (Optional)</Label><Textarea id="note" {...register("note")} /></div>
            </div>

            {/* --- 提交按钮 --- */}
            <div className="flex justify-end gap-2 pt-4"><Button type="button" variant="ghost" onClick={() => navigate(isEditMode ? `/jobs/${id}`: "/jobs")}>Cancel</Button><Button type="submit">{isEditMode ? "Save Changes" : "Create Job"}</Button></div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
};

export default JobForm;