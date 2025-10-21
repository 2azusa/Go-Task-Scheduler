import { useEffect, useState } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api } from "@/lib/api";
import { Script } from "@/types/api";
import { Plus, Trash2, Pencil } from "lucide-react";
import { toast } from "sonner";
import { ScriptEdit } from "@/components/business/ScriptEdit";

const Scripts = () => {
  const [scripts, setScripts] = useState<Script[]>([]);
  const [loading, setLoading] = useState(false);
  const [editingScript, setEditingScript] = useState<Partial<Script> | null>(
    null
  );

  useEffect(() => {
    loadScripts();
  }, []);

  const loadScripts = async () => {
    setLoading(true);
    try {
      const response = await api.searchScripts({ page: 1, pageSize: 100 });
      if (response.code === 200) {
        setScripts(response.data.list || []);
      }
    } catch (error) {
      toast.error("加载脚本列表失败");
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      const response = await api.deleteScripts([id]);
      if (response.code === 200) {
        toast.success("删除成功");
        loadScripts();
      } else {
        toast.error(response.msg || "删除失败");
      }
    } catch (error) {
      toast.error("删除失败");
    }
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-foreground">脚本管理</h1>
            <p className="text-muted-foreground">管理可重用的脚本模板</p>
          </div>
          <Button onClick={() => setEditingScript({})}>
            <Plus className="mr-2 h-4 w-4" />
            创建脚本
          </Button>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>脚本列表</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>脚本名称</TableHead>
                  <TableHead>命令</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center">
                      加载中...
                    </TableCell>
                  </TableRow>
                ) : scripts.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center text-muted-foreground">
                      暂无脚本
                    </TableCell>
                  </TableRow>
                ) : (
                  scripts.map((script) => (
                    <TableRow key={script.id}>
                      <TableCell>{script.id}</TableCell>
                      <TableCell className="font-medium">{script.name}</TableCell>
                      <TableCell className="max-w-md truncate font-mono text-sm">
                        {script.command}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => setEditingScript(script)}
                          >
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() => script.id && handleDelete(script.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
      {editingScript && (
        <ScriptEdit
          script={editingScript}
          onClose={() => setEditingScript(null)}
          onSuccess={() => {
            setEditingScript(null);
            loadScripts();
          }}
        />
      )}
    </MainLayout>
  );
};

export default Scripts;