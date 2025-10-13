'use client';

import { useUser } from '@auth0/nextjs-auth0';
import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { 
  User, 
  Mail, 
  Phone, 
  Building2, 
  CreditCard, 
  Settings,
  Save,
  Loader2,
  AlertCircle,
  CheckCircle,
  Bell,
  Palette,
  Home,
  Receipt
} from 'lucide-react';

interface Subscription {
  id: string;
  status: string;
  priceId?: string;
  currentPeriodEnd?: string;
}

export default function ProfilePage() {
  const { user, isLoading: authLoading } = useUser();
  const router = useRouter();
  
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  
  const [formData, setFormData] = useState({
    fullName: '',
    companyName: '',
    phone: '',
    addressLine1: '',
    addressLine2: '',
    city: '',
    state: '',
    postalCode: '',
    country: 'US',
    emailNotifications: true,
    marketingEmails: false,
    defaultRoomType: 'living_room',
    defaultStyle: 'modern',
  });

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/api/auth/login?returnTo=/profile');
    }
  }, [user, authLoading, router]);

  // Fetch user profile
  useEffect(() => {
    if (user) {
      fetchProfile();
      fetchSubscription();
    }
  }, [user]);

  const fetchProfile = async () => {
    try {
      const res = await fetch('/api/user/profile');
      if (res.ok) {
        const data = await res.json();
        
        // Populate form
        setFormData({
          fullName: data.fullName || '',
          companyName: data.companyName || '',
          phone: data.phone || '',
          addressLine1: data.billingAddress?.line1 || '',
          addressLine2: data.billingAddress?.line2 || '',
          city: data.billingAddress?.city || '',
          state: data.billingAddress?.state || '',
          postalCode: data.billingAddress?.postalCode || '',
          country: data.billingAddress?.country || 'US',
          emailNotifications: data.preferences?.emailNotifications ?? true,
          marketingEmails: data.preferences?.marketingEmails ?? false,
          defaultRoomType: data.preferences?.defaultRoomType || 'living_room',
          defaultStyle: data.preferences?.defaultStyle || 'modern',
        });
      }
    } catch (error) {
      console.error('Failed to fetch profile:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchSubscription = async () => {
    try {
      const res = await fetch('/api/v1/billing/subscriptions');
      if (res.ok) {
        const data = await res.json();
        if (data.items && data.items.length > 0) {
          setSubscription(data.items[0]);
        }
      }
    } catch (error) {
      console.error('Failed to fetch subscription:', error);
    }
  };

  const handleSave = useCallback(async () => {
    setSaving(true);
    setMessage(null);

    try {
      const res = await fetch('/api/user/profile', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          fullName: formData.fullName,
          companyName: formData.companyName,
          phone: formData.phone,
          billingAddress: {
            line1: formData.addressLine1,
            line2: formData.addressLine2,
            city: formData.city,
            state: formData.state,
            postalCode: formData.postalCode,
            country: formData.country,
          },
          preferences: {
            emailNotifications: formData.emailNotifications,
            marketingEmails: formData.marketingEmails,
            defaultRoomType: formData.defaultRoomType,
            defaultStyle: formData.defaultStyle,
          },
        }),
      });

      if (res.ok) {
        setMessage({ type: 'success', text: 'Profile updated successfully!' });
        fetchProfile(); // Refresh
      } else {
        throw new Error('Failed to update profile');
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to update profile. Please try again.' });
    } finally {
      setSaving(false);
      setTimeout(() => setMessage(null), 5000);
    }
  }, [formData]);

  const handleSubscribe = async () => {
    try {
      const res = await fetch('/api/v1/billing/create-checkout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ priceId: 'price_1ABC123' }), // Replace with actual price ID
      });

      if (res.ok) {
        const data = await res.json();
        window.location.href = data.url; // Redirect to Stripe Checkout
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to start checkout. Please try again.' });
    }
  };

  const handleManageBilling = async () => {
    try {
      const res = await fetch('/api/v1/billing/portal', {
        method: 'POST',
      });

      if (res.ok) {
        const data = await res.json();
        window.location.href = data.url; // Redirect to Stripe Customer Portal
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to open billing portal. Please try again.' });
    }
  };

  // Keyboard shortcut: Cmd+Enter (Mac) or Ctrl+Enter (Windows) to save
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault();
        if (!saving) {
          handleSave();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [saving, handleSave]);

  if (authLoading || loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
        <p className="text-gray-600 dark:text-gray-400">Loading your profile...</p>
      </div>
    );
  }

  if (!user) return null;

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Profile Settings</h1>
        <p className="text-gray-600 dark:text-gray-400">
          Manage your account information, billing, and preferences
        </p>
      </div>

      {/* Message Banner */}
      {message && (
        <div className={`rounded-lg border p-4 flex items-start gap-3 ${
          message.type === 'success' 
            ? 'bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800' 
            : 'bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-800'
        }`}>
          {message.type === 'success' ? (
            <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400 mt-0.5" />
          ) : (
            <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" />
          )}
          <p className={message.type === 'success' ? 'text-green-800 dark:text-green-300' : 'text-red-800 dark:text-red-300'}>
            {message.text}
          </p>
        </div>
      )}

      {/* Personal Information */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <User className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Personal Information</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Your basic account details
          </p>
        </div>
        <div className="card-body space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">Full Name</label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="text"
                  value={formData.fullName}
                  onChange={(e) => setFormData({ ...formData, fullName: e.target.value })}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                  placeholder="John Doe"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Email</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="email"
                  value={user.email || ''}
                  disabled
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-slate-800 text-gray-500 cursor-not-allowed"
                />
              </div>
              <p className="text-xs text-gray-500 mt-1">Email is managed by your Auth0 account</p>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Phone Number</label>
              <div className="relative">
                <Phone className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="tel"
                  value={formData.phone}
                  onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                  placeholder="+1 (555) 123-4567"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Business Information */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Business Information</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Company details and billing address
          </p>
        </div>
        <div className="card-body space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Company Name</label>
            <div className="relative">
              <Building2 className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                value={formData.companyName}
                onChange={(e) => setFormData({ ...formData, companyName: e.target.value })}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                placeholder="Acme Real Estate"
              />
            </div>
          </div>

          <div className="space-y-4">
            <label className="block text-sm font-medium">Billing Address</label>
            
            <div>
              <input
                type="text"
                value={formData.addressLine1}
                onChange={(e) => setFormData({ ...formData, addressLine1: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                placeholder="Street address"
              />
            </div>

            <div>
              <input
                type="text"
                value={formData.addressLine2}
                onChange={(e) => setFormData({ ...formData, addressLine2: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                placeholder="Apartment, suite, etc. (optional)"
              />
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="col-span-2">
                <input
                  type="text"
                  value={formData.city}
                  onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                  placeholder="City"
                />
              </div>

              <div>
                <input
                  type="text"
                  value={formData.state}
                  onChange={(e) => setFormData({ ...formData, state: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                  placeholder="State"
                />
              </div>

              <div>
                <input
                  type="text"
                  value={formData.postalCode}
                  onChange={(e) => setFormData({ ...formData, postalCode: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                  placeholder="ZIP"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Payment & Billing */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <CreditCard className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Payment & Billing</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Manage your subscription and payment methods
          </p>
        </div>
        <div className="card-body space-y-4">
          {subscription ? (
            <div className="space-y-4">
              <div className="flex items-center justify-between p-4 bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 rounded-lg">
                <div>
                  <p className="font-medium text-green-900 dark:text-green-300">Active Subscription</p>
                  <p className="text-sm text-green-700 dark:text-green-400 mt-1">
                    Status: {subscription.status}
                  </p>
                  {subscription.currentPeriodEnd && (
                    <p className="text-xs text-green-600 dark:text-green-500 mt-1">
                      Renews on {new Date(subscription.currentPeriodEnd).toLocaleDateString()}
                    </p>
                  )}
                </div>
                <CheckCircle className="h-8 w-8 text-green-600" />
              </div>

              <button
                onClick={handleManageBilling}
                className="w-full btn btn-outline flex items-center justify-center gap-2"
              >
                <Receipt className="h-4 w-4" />
                Manage Subscription & Payment Methods
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <div className="p-6 border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg text-center">
                <CreditCard className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                <p className="text-gray-600 dark:text-gray-400 mb-4">
                  No active subscription
                </p>
                <p className="text-sm text-gray-500 dark:text-gray-500 mb-6">
                  Subscribe to unlock unlimited image staging and priority support
                </p>
                <button
                  onClick={handleSubscribe}
                  className="btn btn-primary inline-flex items-center gap-2"
                >
                  <CreditCard className="h-4 w-4" />
                  Subscribe Now
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
                  <p className="font-medium mb-1">Free Tier</p>
                  <p className="text-2xl font-bold text-blue-600 mb-2">$0<span className="text-sm text-gray-500">/mo</span></p>
                  <ul className="space-y-1 text-gray-600 dark:text-gray-400">
                    <li>✓ 5 images/month</li>
                    <li>✓ Standard processing</li>
                    <li>✓ Email support</li>
                  </ul>
                </div>

                <div className="p-4 border-2 border-blue-600 rounded-lg relative">
                  <div className="absolute -top-3 left-1/2 transform -translate-x-1/2 bg-blue-600 text-white px-3 py-1 rounded-full text-xs font-medium">
                    Popular
                  </div>
                  <p className="font-medium mb-1">Pro</p>
                  <p className="text-2xl font-bold text-blue-600 mb-2">$29<span className="text-sm text-gray-500">/mo</span></p>
                  <ul className="space-y-1 text-gray-600 dark:text-gray-400">
                    <li>✓ 100 images/month</li>
                    <li>✓ Priority processing</li>
                    <li>✓ Chat support</li>
                  </ul>
                </div>

                <div className="p-4 border border-gray-200 dark:border-gray-800 rounded-lg">
                  <p className="font-medium mb-1">Business</p>
                  <p className="text-2xl font-bold text-blue-600 mb-2">$99<span className="text-sm text-gray-500">/mo</span></p>
                  <ul className="space-y-1 text-gray-600 dark:text-gray-400">
                    <li>✓ 500 images/month</li>
                    <li>✓ Fastest processing</li>
                    <li>✓ Priority support</li>
                  </ul>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Preferences */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <Settings className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Preferences</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Customize your experience
          </p>
        </div>
        <div className="card-body space-y-6">
          {/* Notifications */}
          <div className="space-y-4">
            <h3 className="font-medium flex items-center gap-2">
              <Bell className="h-4 w-4" />
              Notifications
            </h3>
            
            <label className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-800 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-slate-800/50">
              <div>
                <p className="font-medium">Email Notifications</p>
                <p className="text-sm text-gray-500">Receive updates about your image processing</p>
              </div>
              <input
                type="checkbox"
                checked={formData.emailNotifications}
                onChange={(e) => setFormData({ ...formData, emailNotifications: e.target.checked })}
                className="h-5 w-5 text-blue-600 rounded focus:ring-2 focus:ring-blue-500"
              />
            </label>

            <label className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-800 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-slate-800/50">
              <div>
                <p className="font-medium">Marketing Emails</p>
                <p className="text-sm text-gray-500">Receive news, tips, and special offers</p>
              </div>
              <input
                type="checkbox"
                checked={formData.marketingEmails}
                onChange={(e) => setFormData({ ...formData, marketingEmails: e.target.checked })}
                className="h-5 w-5 text-blue-600 rounded focus:ring-2 focus:ring-blue-500"
              />
            </label>
          </div>

          {/* Default Settings */}
          <div className="space-y-4">
            <h3 className="font-medium flex items-center gap-2">
              <Palette className="h-4 w-4" />
              Default Staging Settings
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2 flex items-center gap-2">
                  <Home className="h-4 w-4" />
                  Default Room Type
                </label>
                <select
                  value={formData.defaultRoomType}
                  onChange={(e) => setFormData({ ...formData, defaultRoomType: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                >
                  <option value="living_room">Living Room</option>
                  <option value="bedroom">Bedroom</option>
                  <option value="kitchen">Kitchen</option>
                  <option value="bathroom">Bathroom</option>
                  <option value="dining_room">Dining Room</option>
                  <option value="office">Office</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2 flex items-center gap-2">
                  <Palette className="h-4 w-4" />
                  Default Style
                </label>
                <select
                  value={formData.defaultStyle}
                  onChange={(e) => setFormData({ ...formData, defaultStyle: e.target.value })}
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900"
                >
                  <option value="modern">Modern</option>
                  <option value="contemporary">Contemporary</option>
                  <option value="traditional">Traditional</option>
                  <option value="scandinavian">Scandinavian</option>
                  <option value="industrial">Industrial</option>
                  <option value="bohemian">Bohemian</option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Spacer for sticky button bar */}
      <div className="h-24" />

      {/* Sticky Save Button Bar */}
      <div className="fixed bottom-0 left-0 right-0 z-40 border-t border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-slate-950/80 backdrop-blur-xl supports-[backdrop-filter]:bg-white/60 dark:supports-[backdrop-filter]:bg-slate-950/60">
        <div className="container max-w-5xl mx-auto py-4 flex justify-between items-center">
          {/* Keyboard shortcut hint */}
          <div className="hidden sm:flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
            <kbd className="px-2 py-1 bg-gray-100 dark:bg-slate-800 border border-gray-300 dark:border-gray-700 rounded text-xs font-mono">
              {typeof navigator !== 'undefined' && navigator.platform.includes('Mac') ? '⌘' : 'Ctrl'}
            </kbd>
            <span>+</span>
            <kbd className="px-2 py-1 bg-gray-100 dark:bg-slate-800 border border-gray-300 dark:border-gray-700 rounded text-xs font-mono">
              Enter
            </kbd>
            <span>to save</span>
          </div>
          
          <div className="flex gap-3 ml-auto">
            <button
              onClick={() => router.push('/')}
              className="btn btn-outline"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={saving}
              className="btn btn-primary flex items-center gap-2"
            >
              {saving ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4" />
                  Save Changes
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
