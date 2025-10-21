import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { toast } from "sonner";
import { api } from "@/lib/api";
import { Calendar } from "lucide-react";

const Login = () => {
  const navigate = useNavigate(); // 用于页面跳转
  // 登录表单状态
  const [loginForm, setLoginForm] = useState({ username: "", password: "" });
  // 注册表单状态
  const [registerForm, setRegisterForm] = useState({
    username: "",
    password: "",
    email: "",
  });
  const [loading, setLoading] = useState(false);

  // 登录提交处理函数
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault(); // 阻止表单默认提交
    setLoading(true); // 开始加载
    try {
      const response = await api.login(loginForm);
      if (response.code === 200) {
        localStorage.setItem("token", response.data.token);
        toast.success("登录成功");
        navigate("/dashboard");
      } else {
        toast.error(response.msg || "登录失败");
      }
    } catch (error) {
      toast.error("网络错误，请稍后重试");
    } finally {
      setLoading(false);
    }
  };

  // 注册提交处理函数
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      // 
      const response = await api.register(registerForm);
      if (response.code === 200) {
        toast.success("注册成功，请登录");
        setRegisterForm({ username: "", password: "", email: "" });
      } else {
        toast.error(response.msg || "注册失败");
      }
    } catch (error) {
      toast.error("网络错误，请稍后重试");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <div className="mb-4 flex justify-center">
            <div className="rounded-lg bg-primary p-3">
              <Calendar className="h-8 w-8 text-primary-foreground" />
            </div>
          </div>
          <h1 className="text-3xl font-bold text-foreground">分布式任务调度平台</h1>
          <p className="mt-2 text-muted-foreground">
            集中管理和调度您的定时任务
          </p>
        </div>

        <Tabs defaultValue="login" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="login">登录</TabsTrigger>
            <TabsTrigger value="register">注册</TabsTrigger>
          </TabsList>

          <TabsContent value="login">
            <Card>
              <CardHeader>
                <CardTitle>欢迎回来</CardTitle>
                <CardDescription>使用您的账号登录系统</CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleLogin} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="login-username">用户名</Label>
                    <Input
                      id="login-username"
                      placeholder="请输入用户名"
                      value={loginForm.username}
                      onChange={(e) =>
                        setLoginForm({ ...loginForm, username: e.target.value })
                      }
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="login-password">密码</Label>
                    <Input
                      id="login-password"
                      type="password"
                      placeholder="请输入密码"
                      value={loginForm.password}
                      onChange={(e) =>
                        setLoginForm({ ...loginForm, password: e.target.value })
                      }
                      required
                    />
                  </div>
                  <Button type="submit" className="w-full" disabled={loading}>
                    {loading ? "登录中..." : "登录"}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="register">
            <Card>
              <CardHeader>
                <CardTitle>创建账号</CardTitle>
                <CardDescription>注册新的系统账号</CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleRegister} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="register-username">用户名</Label>
                    <Input
                      id="register-username"
                      placeholder="请输入用户名 (2-20个字符)"
                      value={registerForm.username}
                      onChange={(e) =>
                        setRegisterForm({ ...registerForm, username: e.target.value })
                      }
                      required
                      minLength={2}
                      maxLength={20}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="register-email">邮箱</Label>
                    <Input
                      id="register-email"
                      type="email"
                      placeholder="请输入邮箱"
                      value={registerForm.email}
                      onChange={(e) =>
                        setRegisterForm({ ...registerForm, email: e.target.value })
                      }
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="register-password">密码</Label>
                    <Input
                      id="register-password"
                      type="password"
                      placeholder="请输入密码 (4-20个字符)"
                      value={registerForm.password}
                      onChange={(e) =>
                        setRegisterForm({ ...registerForm, password: e.target.value })
                      }
                      required
                      minLength={4}
                      maxLength={20}
                    />
                  </div>
                  <Button type="submit" className="w-full" disabled={loading}>
                    {loading ? "注册中..." : "注册"}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
};

export default Login;
