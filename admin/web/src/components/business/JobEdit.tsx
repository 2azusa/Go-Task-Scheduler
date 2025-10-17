
import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Job } from "@/types/api";
import { api } from "@/lib/api";
import { toast } from "sonner";

interface JobEditProps {
  job: Partial<Job> | null;
  onClose: () => void;
  onSuccess: () => void;
}

export const JobEdit = ({ job, onClose, onSuccess }: JobEditProps) => {
  const [formData, setFormData] = useState<Partial<Job>>({});

  useEffect(() => {
    if (job) {
      setFormData(job);
    } else {
      setFormData({
        job_type: 1,
        status: 1,
      });
    }
  }, [job]);

  const handleChange = (name: keyof Job, value: any) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async () => {
    try {
      const apiCall = job?.id ? api.updateJob : api.createJob;
      const response = await apiCall(formData);
      if (response.code === 200) {
        toast.success(job?.id ? "更新成功" : "创建成功");
        onSuccess();
      } else {
        toast.error(response.msg || (job?.id ? "更新失败" : "创建失败"));
      }
    } catch (error) {
      toast.error(job?.id ? "更新失败" : "创建失败");
    }
  };

  return (
    <Dialog open={!!job} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{job?.id ? "编辑任务" : "创建任务"}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              任务名称
            </Label>
            <Input
              id="name"
              value={formData.name || ""}
              onChange={(e) => handleChange("name", e.target.value)}
              className="col-span-3"
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="job_type" className="text-right">
              任务类型
            </Label>
            <Select
              value={String(formData.job_type)}
              onValueChange={(value) => handleChange("job_type", parseInt(value))}
            >
              <SelectTrigger className="col-span-3">
                <SelectValue placeholder="选择任务类型" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">命令任务</SelectItem>
                <SelectItem value="2">HTTP任务</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="spec" className="text-right">
              Cron表达式
            </Label>
            <Input
              id="spec"
              value={formData.spec || ""}
              onChange={(e) => handleChange("spec", e.target.value)}
              className="col-span-3"
            />
          </div>
          {formData.job_type === 1 ? (
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="command" className="text-right">
                命令
              </Label>
              <Input
                id="command"
                value={formData.command || ""}
                onChange={(e) => handleChange("command", e.target.value)}
                className="col-span-3"
              />
            </div>
          ) : (
            <>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="http_url" className="text-right">
                  URL
                </Label>
                <Input
                  id="http_url"
                  value={formData.http_url || ""}
                  onChange={(e) => handleChange("http_url", e.target.value)}
                  className="col-span-3"
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="http_method" className="text-right">
                  Method
                </Label>
                <Select
                  value={formData.http_method || "GET"}
                  onValueChange={(value) => handleChange("http_method", value)}
                >
                  <SelectTrigger className="col-span-3">
                    <SelectValue placeholder="选择HTTP方法" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="GET">GET</SelectItem>
                    <SelectItem value="POST">POST</SelectItem>
                    <SelectItem value="PUT">PUT</SelectItem>
                    <SelectItem value="DELETE">DELETE</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button onClick={handleSubmit}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
