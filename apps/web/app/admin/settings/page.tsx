"use client";

import { useState, useEffect } from "react";
import { useAuth0 } from "@auth0/auth0-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { AlertCircle, CheckCircle2, Settings2 } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";

interface ModelInfo {
  id: string;
  name: string;
  description: string;
  version: string;
  is_active: boolean;
}

export default function AdminSettingsPage() {
  const { getAccessTokenSilently } = useAuth0();
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    fetchModels();
  }, []);

  const fetchModels = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const token = await getAccessTokenSilently();
      const response = await fetch("/api/v1/admin/models", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to fetch models");
      }

      const data = await response.json();
      setModels(data.models || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setLoading(false);
    }
  };

  const updateActiveModel = async (modelID: string) => {
    try {
      setUpdating(modelID);
      setError(null);
      setSuccessMessage(null);

      const token = await getAccessTokenSilently();
      const response = await fetch("/api/v1/admin/models/active", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ value: modelID }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to update active model");
      }

      setSuccessMessage("Active model updated successfully!");
      
      // Refresh models list
      await fetchModels();

      // Clear success message after 3 seconds
      setTimeout(() => setSuccessMessage(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setUpdating(null);
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-center min-h-[400px]">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading models...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Settings2 className="h-8 w-8 text-gray-700" />
            <h1 className="text-3xl font-bold">Admin Settings</h1>
          </div>
          <p className="text-gray-600">
            Configure the active AI model for virtual staging
          </p>
        </div>

        {/* Error Alert */}
        {error && (
          <Alert variant="destructive" className="mb-6">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Success Alert */}
        {successMessage && (
          <Alert className="mb-6 border-green-500 bg-green-50">
            <CheckCircle2 className="h-4 w-4 text-green-600" />
            <AlertDescription className="text-green-800">
              {successMessage}
            </AlertDescription>
          </Alert>
        )}

        {/* Models Section */}
        <Card>
          <CardHeader>
            <CardTitle>AI Models</CardTitle>
            <CardDescription>
              Select which AI model to use for processing staging requests
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {models.map((model) => (
                <div
                  key={model.id}
                  className={`border rounded-lg p-4 transition-all ${
                    model.is_active
                      ? "border-blue-500 bg-blue-50"
                      : "border-gray-200 hover:border-gray-300"
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <h3 className="text-lg font-semibold">{model.name}</h3>
                        {model.is_active && (
                          <Badge className="bg-blue-600">Active</Badge>
                        )}
                        <Badge variant="outline">{model.version}</Badge>
                      </div>
                      <p className="text-sm text-gray-600 mb-3">
                        {model.description}
                      </p>
                      <p className="text-xs text-gray-500 font-mono">
                        {model.id}
                      </p>
                    </div>
                    <div className="ml-4">
                      {!model.is_active && (
                        <Button
                          onClick={() => updateActiveModel(model.id)}
                          disabled={updating === model.id}
                          variant="outline"
                        >
                          {updating === model.id ? "Activating..." : "Activate"}
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Info Box */}
        <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <p className="text-sm text-blue-800">
            <strong>Note:</strong> Changing the active model will affect all new
            staging requests. Existing jobs in progress will continue using their
            original model.
          </p>
        </div>
      </div>
    </div>
  );
}
