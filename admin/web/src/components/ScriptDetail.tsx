import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { api } from "@/lib/api";
import { Script } from "@/types/api";
import { toast } from "sonner";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Activity, ArrowLeft, Edit } from "lucide-react";
import { format } from 'date-fns';
import { Label } from "@radix-ui/react-label";

const ScriptDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const scriptId = id ? parseInt(id, 10) : 0;

  const [script, setScript] = useState<Script | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (scriptId > 0) {
      api.findScriptById({ id: scriptId })
        .then(setScript)
        .catch(error => {
          toast.error(`Failed to load script: ${error.message}`);
          navigate("/scripts");
        })
        .finally(() => setIsLoading(false));
    }
  }, [scriptId, navigate]);

  const formatDate = (ts: number) => ts ? format(new Date(ts * 1000), 'yyyy-MM-dd HH:mm:ss') : 'N/A';

  if (isLoading) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  if (!script) {
    return (
      <div className="container mx-auto p-6 text-center">
        <h2 className="text-xl text-muted-foreground">Script not found.</h2>
        <Button variant="outline" onClick={() => navigate("/scripts")} className="mt-4"><ArrowLeft className="mr-2 h-4 w-4" /> Back to Scripts</Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Button variant="outline" onClick={() => navigate("/scripts")} className="mb-4"><ArrowLeft className="mr-2 h-4 w-4" /> Back to Scripts List</Button>
      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
            <div>
              <CardTitle className="text-2xl">{script.name}</CardTitle>
              <CardDescription>Last updated: {formatDate(script.updated)}</CardDescription>
            </div>
            <Button variant="outline" onClick={() => navigate(`/scripts/${script.id}/edit`)}>
              <Edit className="mr-2 h-4 w-4" /> Edit
            </Button>
          </div>
        </CardHeader>
        <CardContent>
            <Label className="text-muted-foreground">Script Content</Label>
            <pre className="mt-2 p-4 bg-muted rounded-md text-sm font-mono overflow-auto max-h-[60vh]">
                <code>{script.command}</code>
            </pre>
        </CardContent>
      </Card>
    </div>
  );
};

export default ScriptDetail;