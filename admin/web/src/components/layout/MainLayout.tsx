import { ReactNode } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  Calendar,
  Server,
  FileCode,
  Users,
  ListOrdered,
  Settings,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { UserNav } from "./UserNav";

interface MainLayoutProps {
  children: ReactNode;
}

const menuItems = [
  { path: "/dashboard", icon: LayoutDashboard, label: "仪表盘" },
  { path: "/jobs", icon: Calendar, label: "任务管理" },
  { path: "/nodes", icon: Server, label: "节点管理" },
  { path: "/scripts", icon: FileCode, label: "脚本管理" },
  { path: "/users", icon: Users, label: "用户管理" },
  { path: "/logs", icon: ListOrdered, label: "任务日志" },
  { path: "/settings", icon: Settings, label: "系统信息" },
];

export const MainLayout = ({ children }: MainLayoutProps) => {
  const location = useLocation();
  const navigate = useNavigate();

  return (
    <div className="flex h-screen bg-background">
      {/* Sidebar */}
      <aside className="w-64 border-r border-border bg-card">
        <div className="flex h-16 items-center justify-center border-b border-border">
          <h1 className="text-xl font-bold text-primary">任务调度平台</h1>
        </div>
        <nav className="flex flex-col gap-1 p-4">
          {menuItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.path;
            return (
              <Link key={item.path} to={item.path}>
                <Button
                  variant={isActive ? "default" : "ghost"}
                  className={cn(
                    "w-full justify-start gap-3",
                    !isActive && "text-muted-foreground hover:text-foreground"
                  )}
                >
                  <Icon className="h-5 w-5" />
                  {item.label}
                </Button>
              </Link>
            );
          })}
        </nav>
      </aside>

      {/* Main Content */}
      <div className="flex flex-1 flex-col overflow-auto">
        <header className="flex h-16 items-center justify-end border-b border-border px-8">
          <UserNav />
        </header>
        <main className="flex-1 p-8">{children}</main>
      </div>
    </div>
  );
};
