import { useEffect, useState, useCallback } from "react"; // 1. Import useCallback
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { Job, JobStatus, JobType } from "@/types/api";
import { Activity, Search, Plus, Play, Trash2, StopCircle, RotateCcw } from "lucide-react";

const Jobs = () => {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const navigate = useNavigate();
  
  const [searchQuery, setSearchQuery] = useState("");
  const [filterStatus, setFilterStatus] = useState<string>("all");
  const [filterType, setFilterType] = useState<string>("all");

  const fetchJobs = useCallback(async (resetPage = false) => {
    setIsLoading(true);
    try {
      const currentPage = resetPage ? 1 : page;
      if (resetPage) setPage(1);

      const jobsData = await api.searchJobs({
        page: currentPage,
        pageSize: 10,
        name: searchQuery || undefined,
        status: filterStatus !== "all" ? Number(filterStatus) as JobStatus : undefined,
        type: filterType !== "all" ? Number(filterType) as JobType : undefined,
      });
      
      setJobs(jobsData.list || []);
      setTotal(jobsData.total);

    } catch (error) {
      if (error instanceof Error) toast.error(error.message);
      else toast.error("Failed to load jobs due to an unknown error");
    } finally {
      setIsLoading(false);
    }
  }, [page, searchQuery, filterStatus, filterType]);

  useEffect(() => {
    fetchJobs();
  }, [fetchJobs]);

  const handleSearch = () => {
    fetchJobs(true);
  };

  const handleResetFilters = () => {
    setSearchQuery("");
    setFilterStatus("all");
    setFilterType("all");
    fetchJobs(true);
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') handleSearch();
  };

  const handleExecute = async (job: Job) => {
    if (window.confirm(`Are you sure you want to execute "${job.name}" immediately?`)) {
      try {
        await api.executeJobOnce({ jobId: job.id, nodeUuid: job.runOn });
        toast.success(`Job "${job.name}" has been triggered.`);
      } catch (error) {
        if (error instanceof Error) toast.error(error.message);
      }
    }
  };

  const handleKill = async (job: Job) => {
     if (window.confirm(`Are you sure you want to kill "${job.name}"?`)) {
      try {
        await api.killJob({ jobId: job.id, nodeUuid: job.runOn });
        toast.success(`Request to kill job "${job.name}" sent.`);
      } catch (error) {
        if (error instanceof Error) toast.error(error.message);
      }
    }
  };
  
  const handleDelete = async (jobId: number, jobName: string) => {
     if (window.confirm(`Are you sure you want to delete "${jobName}"?`)) {
      try {
        await api.deleteJobs({ ids: [jobId] });
        toast.success(`Job "${jobName}" deleted.`);
        fetchJobs();
      } catch (error) {
        if (error instanceof Error) toast.error(error.message);
      }
    }
  };

  const getStatusBadge = (status: JobStatus) => {
    return status === JobStatus.JobStatusAssigned ? (
      <Badge className="bg-success/10 text-success border-success/20 hover:bg-success/20">Assigned</Badge>
    ) : (
      <Badge variant="secondary">Not Assigned</Badge>
    );
  };

  const getTypeBadge = (type: JobType) => {
    return type === JobType.JobTypeCmd ? <Badge variant="outline">Shell</Badge> : <Badge variant="outline">HTTP</Badge>;
  };

  if (isLoading && jobs.length === 0) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold mb-2">Jobs</h1>
        <p className="text-muted-foreground">Manage and monitor your scheduled tasks</p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div>
              <CardTitle>All Jobs</CardTitle>
              <CardDescription>View, filter, and manage scheduled jobs</CardDescription>
            </div>
            <Button className="gap-2 w-full sm:w-auto" onClick={() => navigate("/jobs/new")}>
              <Plus className="h-4 w-4" /> Create Job
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div className="relative lg:col-span-2">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search by job name..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={handleKeyDown}
                className="pl-9"
              />
            </div>
            <Select value={filterStatus} onValueChange={setFilterStatus}>
              <SelectTrigger><SelectValue placeholder="Filter by Status" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Statuses</SelectItem>
                <SelectItem value={String(JobStatus.JobStatusAssigned)}>Assigned</SelectItem>
                <SelectItem value={String(JobStatus.JobStatusNotAssigned)}>Not Assigned</SelectItem>
              </SelectContent>
            </Select>
             <Select value={filterType} onValueChange={setFilterType}>
              <SelectTrigger><SelectValue placeholder="Filter by Type" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value={String(JobType.JobTypeCmd)}>Shell</SelectItem>
                <SelectItem value={String(JobType.JobTypeHttp)}>HTTP</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex justify-end gap-2 mb-4">
              <Button variant="outline" onClick={handleResetFilters} className="gap-2"><RotateCcw className="h-4 w-4"/> Reset</Button>
              <Button onClick={handleSearch} className="gap-2"><Search className="h-4 w-4"/> Search</Button>
          </div>

          <div className="border rounded-lg">
            <TooltipProvider>
              <Table>
                 <TableHeader>
                    <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Type</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Schedule</TableHead>
                        <TableHead>Node</TableHead>
                        <TableHead>Actions</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                {isLoading && jobs.length === 0 ? (
                  <TableRow><TableCell colSpan={6} className="text-center py-8"><Activity className="h-6 w-6 animate-spin text-primary mx-auto" /></TableCell></TableRow>
                ) : !isLoading && jobs.length === 0 ? (
                  <TableRow><TableCell colSpan={6} className="text-center py-8 text-muted-foreground">No jobs found.</TableCell></TableRow>
                ) : (
                  jobs.map((job) => (
                    <TableRow key={job.id}>
                      <TableCell className="font-medium">
                        <a onClick={() => navigate(`/jobs/${job.id}`)} className="hover:underline cursor-pointer text-primary">{job.name}</a>
                      </TableCell>
                      <TableCell>{getTypeBadge(job.jobType)}</TableCell>
                      <TableCell>{getStatusBadge(job.status)}</TableCell>
                      <TableCell className="font-mono text-sm">{job.spec}</TableCell>
                      <TableCell>{job.hostname || <span className="text-muted-foreground text-sm">-</span>}</TableCell>
                      <TableCell>
                         <div className="flex gap-1">
                            <Tooltip><TooltipTrigger asChild><Button size="icon" variant="ghost" className="h-8 w-8" disabled={!job.runOn} onClick={() => handleExecute(job)}><Play className="h-4 w-4" /></Button></TooltipTrigger><TooltipContent>{job.runOn ? <p>Execute Once</p> : <p>Job must be assigned</p>}</TooltipContent></Tooltip>
                            <Tooltip><TooltipTrigger asChild><Button size="icon" variant="ghost" className="h-8 w-8 text-orange-500" disabled={!job.runOn} onClick={() => handleKill(job)}><StopCircle className="h-4 w-4" /></Button></TooltipTrigger><TooltipContent>{job.runOn ? <p>Kill Job</p> : <p>Job must be assigned</p>}</TooltipContent></Tooltip>
                             <Tooltip><TooltipTrigger asChild><Button size="icon" variant="ghost" className="h-8 w-8 text-destructive" onClick={() => handleDelete(job.id, job.name)}><Trash2 className="h-4 w-4" /></Button></TooltipTrigger><TooltipContent><p>Delete Job</p></TooltipContent></Tooltip>
                          </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
                </TableBody>
              </Table>
            </TooltipProvider>
          </div>

          {total > 10 && (
            <div className="flex items-center justify-between mt-4">
              <p className="text-sm text-muted-foreground">Showing {Math.min((page - 1) * 10 + 1, total)} to {Math.min(page * 10, total)} of {total} jobs</p>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" onClick={() => setPage(p => p - 1)} disabled={page === 1}>Previous</Button>
                <Button variant="outline" size="sm" onClick={() => setPage(p => p + 1)} disabled={page * 10 >= total}>Next</Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default Jobs;