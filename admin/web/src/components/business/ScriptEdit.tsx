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
import { Script } from "@/types/api";
import { api } from "@/lib/api";
import { toast } from "sonner";

interface ScriptEditProps {
  script: Partial<Script> | null;
  onClose: () => void;
  onSuccess: () => void;
}

export const ScriptEdit = ({ script, onClose, onSuccess }: ScriptEditProps) => {
  const [formData, setFormData] = useState<Partial<Script>>({});

  useEffect(() => {
    if (script) {
      setFormData(script);
    } else {
      setFormData({});
    }
  }, [script]);

  const handleChange = (name: keyof Script, value: any) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async () => {
    try {
      const apiCall = script?.id ? api.updateScript : api.createScript;
      const response = await apiCall(formData as Script);
      if (response.code === 200) {
        toast.success(script?.id ? "更新成功" : "创建成功");
        onSuccess();
      } else {
        toast.error(response.msg || (script?.id ? "更新失败" : "创建失败"));
      }
    } catch (error) {
      toast.error(script?.id ? "更新失败" : "创建失败");
    }
  };

  return (
    <Dialog open={!!script} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{script?.id ? "编辑脚本" : "创建脚本"}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="name" className="text-right">
              脚本名称
            </Label>
            <Input
              id="name"
              value={formData.name || ""}
              onChange={(e) => handleChange("name", e.target.value)}
              className="col-span-3"
            />
          </div>
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
