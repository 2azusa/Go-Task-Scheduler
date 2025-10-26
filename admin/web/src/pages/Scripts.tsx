import { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { Script } from "@/types/api";
import { Activity, Search, Plus, Trash2, Pencil } from "lucide-react";

const Scripts = () => {
  const [scripts, setScripts] = useState<Script[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const navigate = useNavigate();

  const fetchScripts = useCallback(async () => {
    setIsLoading(true);
    try {
      const scriptsData = await api.searchScripts({ page, pageSize: 12, name: searchQuery || undefined });
      setScripts(scriptsData.list || []);
      setTotal(scriptsData.total);
    } catch (error) {
      if (error instanceof Error) toast.error(error.message);
      else toast.error("Failed to load scripts");
    } finally {
      setIsLoading(false);
    }
  }, [page, searchQuery]);

  useEffect(() => {
    fetchScripts();
  }, [fetchScripts]);

  const handleSearch = () => {
    setPage(1);
    fetchScripts();
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') handleSearch();
  };

  const handleDelete = async (scriptId: number, scriptName: string) => {
    if (window.confirm(`Delete script "${scriptName}"?`)) {
      try {
        await api.deleteScripts({ ids: [scriptId] });
        toast.success(`Script "${scriptName}" deleted.`);
        fetchScripts();
      } catch (error) {
        if (error instanceof Error) toast.error(error.message);
      }
    }
  };

  const formatDate = (timestamp: number) => {
    if (!timestamp) return 'N/A';
    return new Date(timestamp * 1000).toLocaleDateString();
  };

  if (isLoading && page === 1 && scripts.length === 0) {
    return <div className="flex items-center justify-center h-full"><Activity className="h-8 w-8 animate-spin text-primary" /></div>;
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold mb-2">Scripts</h1>
        <p className="text-muted-foreground">Manage reusable script templates for your jobs</p>
      </div>
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div><CardTitle>Script Library</CardTitle><CardDescription>Browse and manage reusable scripts.</CardDescription></div>
            <Button className="gap-2 w-full sm:w-auto" onClick={() => navigate("/scripts/new")}>
              <Plus className="h-4 w-4" /> New Script
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 mb-6">
             <div className="relative flex-grow"><Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" /><Input placeholder="Search by script name..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} onKeyDown={handleKeyDown} className="pl-9" /></div>
            <Button onClick={handleSearch}>Search</Button>
          </div>

          {isLoading ? (<div className="text-center py-12"><Activity className="h-6 w-6 animate-spin text-primary mx-auto" /></div>) :
           scripts.length === 0 ? (<div className="text-center py-12"><p className="text-muted-foreground">No scripts found.</p></div>) :
           (<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {scripts.map((script) => (
                <Card key={script.id} className="hover:border-primary/50 transition-colors flex flex-col">
                  <CardHeader
                    className="cursor-pointer"
                    onClick={() => navigate(`/scripts/${script.id}`)}
                  >
                    <div className="flex items-start justify-between"><CardTitle className="text-base break-all">{script.name}</CardTitle><Badge variant="outline" className="text-xs shrink-0 ml-2">#{script.id}</Badge></div>
                    <CardDescription className="text-xs">Updated: {formatDate(script.updated)}</CardDescription>
                  </CardHeader>
                  <CardContent className="flex-grow flex flex-col justify-between">
                    <pre className="text-xs bg-secondary p-3 rounded-lg overflow-x-auto max-h-24 cursor-pointer" onClick={() => navigate(`/scripts/${script.id}`)}>
                      <code>{script.command.substring(0, 100)}{script.command.length > 100 && "..."}</code>
                    </pre>
                    <div className="flex gap-2 mt-4">
                      <Button size="sm" variant="outline" className="flex-1" onClick={() => navigate(`/scripts/${script.id}/edit`)}>
                        <Pencil className="h-3 w-3 mr-1" /> Edit
                      </Button>
                      <Button size="icon" variant="ghost" className="text-destructive hover:text-destructive h-9 w-9" onClick={() => handleDelete(script.id, script.name)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>)}

          {total > 12 && (<div className="flex items-center justify-between mt-6 pt-6 border-t border-border"><p className="text-sm text-muted-foreground">Showing {Math.min((page - 1) * 12 + 1, total)} to {Math.min(page * 12, total)} of {total} scripts</p><div className="flex gap-2"><Button variant="outline" size="sm" onClick={() => setPage(page - 1)} disabled={page === 1}>Previous</Button><Button variant="outline" size="sm" onClick={() => setPage(page + 1)} disabled={page * 12 >= total}>Next</Button></div></div>)}
        </CardContent>
      </Card>
    </div>
  );
};

export default Scripts;