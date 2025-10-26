import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { DashboardLayout } from "./components/DashboardLayout";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import Jobs from "./pages/Jobs";
import Nodes from "./pages/Nodes";
import JobLogs from "./pages/JobLogs";
import Scripts from "./pages/Scripts";
import NotFound from "./pages/NotFound";

import JobDetail from "./components/JobDetail";
import JobForm from "./components/JobForm";

import NodeDetail from "./components/NodeDetail";

import ScriptForm from "./components/ScriptForm";
import ScriptDetail from "./components/ScriptDetail";

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <TooltipProvider>
      <Toaster />
      <Sonner />
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="/login" element={<Login />} />
          <Route element={<DashboardLayout />}>
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/jobs" element={<Jobs />} />
            <Route path="/jobs/new" element={<JobForm />} />
            <Route path="/jobs/:id" element={<JobDetail />} />
            <Route path="/jobs/:id/edit" element={<JobForm />} />
            <Route path="/scripts/new" element={<ScriptForm />} />
            <Route path="/scripts/:id" element={<ScriptDetail />} /> 
            <Route path="/scripts/:id/edit" element={<ScriptForm />} />
            <Route path="/nodes" element={<Nodes />} />
            <Route path="/nodes/:uuid" element={<NodeDetail />} />
            <Route path="/logs" element={<JobLogs />} />
            <Route path="/scripts" element={<Scripts />} />
            <Route path="/scripts/:id/edit" element={<ScriptForm />} />
          </Route>
          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </TooltipProvider>
  </QueryClientProvider>
);

export default App;
