import { useEffect, useState } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";
import { User } from "@/types/api";
import { toast } from "sonner";

const Profile = () => {
  const [user, setUser] = useState<User | null>(null);
  const [password, setPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");

  useEffect(() => {
    loadUser();
  }, []);

  const loadUser = async () => {
    try {
      const response = await api.getCurrentUser();
      if (response.code === 200) {
        setUser(response.data);
      }
    } catch (error) {
      toast.error("加载用户信息失败");
    }
  };

  const handlePasswordChange = async () => {
    if (!password || !newPassword) {
      toast.error("请输入当前密码和新密码");
      return;
    }
    try {
      const response = await api.updatePassword(password, newPassword);
      if (response.code === 200) {
        toast.success("密码修改成功");
        setPassword("");
        setNewPassword("");
      } else {
        toast.error(response.msg || "密码修改失败");
      }
    } catch (error) {
      toast.error("密码修改失败");
    }
  };

  return (
    <MainLayout>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">个人信息</h1>
          <p className="text-muted-foreground">查看和编辑您的个人信息</p>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>基本信息</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label>用户名</Label>
                <p className="text-lg font-semibold">{user?.username}</p>
              </div>
              <div>
                <Label>角色</Label>
                <p className="text-lg">{user?.role === 2 ? "管理员" : "普通用户"}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>修改密码</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="current-password">当前密码</Label>
                <Input
                  id="current-password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="new-password">新密码</Label>
                <Input
                  id="new-password"
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                />
              </div>
              <Button onClick={handlePasswordChange}>保存</Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </MainLayout>
  );
};

export default Profile;
