import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { api } from "@/lib/api";
import { ScriptUpdate } from "@/types/api";
import { toast } from "sonner";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Activity, ArrowLeft } from "lucide-react";

const ScriptForm = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isEditMode = !!id;
  const [isLoading, setIsLoading] = useState(isEditMode);

  const { register, handleSubmit, reset, formState: { errors } } = useForm<ScriptUpdate>();

  useEffect(() => {
    if (isEditMode) {
      const scriptId = parseInt(id, 10);
      api.findScriptById({ id: scriptId })
        .then(scriptData => {
          reset(scriptData);
        })
        .catch(error => {
          toast.error(`Failed to load script data: ${error.message}`);
          navigate("/scripts");
        })
        .finally(() => setIsLoading(false));
    }
  }, [id, isEditMode, navigate, reset]);

  const onSubmit = async (data: ScriptUpdate) => {
    const payload = {
      ...data,
      id: isEditMode ? parseInt(id, 10) : undefined,
    };

    try {
      await api.saveScript(payload);
      toast.success(isEditMode ? "Script updated successfully!" : "Script created successfully!");
      navigate("/scripts");
    } catch (error) {
      if (error instanceof Error) toast.error(error.message);
      else toast.error("An unknown error occurred.");
    }
  };

  if (isLoading) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
       <Button variant="outline" onClick={() => navigate("/scripts")} className="mb-4">
        <ArrowLeft className="mr-2 h-4 w-4" /> Back to Scripts
      </Button>
      <Card>
        <CardHeader>
          <CardTitle>{isEditMode ? "Edit Script" : "Create New Script"}</CardTitle>
          <CardDescription>
            {isEditMode ? "Update the script name and content." : "Provide a name and the command content for the new script."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            <div>
              <Label htmlFor="name">Script Name</Label>
              <Input 
                id="name" 
                {...register("name", { required: "Script name is required" })}
                placeholder="e.g., 'Backup Database Script'"
              />
              {errors.name && <p className="text-destructive text-sm mt-1">{errors.name.message}</p>}
            </div>
            <div>
              <Label htmlFor="command">Script Content</Label>
              <Textarea 
                id="command"
                {...register("command", { required: "Script content cannot be empty" })}
                rows={15}
                placeholder="#!/bin/bash&#10;echo 'Hello, World!'"
                className="font-mono"
              />
              {errors.command && <p className="text-destructive text-sm mt-1">{errors.command.message}</p>}
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="ghost" onClick={() => navigate("/scripts")}>Cancel</Button>
              <Button type="submit">
                {isEditMode ? "Save Changes" : "Create Script"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
};

export default ScriptForm;