export type BackendProfile = {
  id: string
  email?: string | null
  full_name?: string | null
  company_name?: string | null
  phone?: string | null
  billing_address?: {
    line1?: string
    line2?: string
    city?: string
    state?: string
    postal_code?: string
    country?: string
  } | null
  profile_photo_url?: string | null
  preferences?: {
    email_notifications?: boolean
    marketing_emails?: boolean
    default_room_type?: string
    default_style?: string
  } | null
  role: string
  stripe_customer_id?: string | null
  created_at: string
  updated_at: string
}

export type ProfileFormData = {
  fullName: string
  companyName: string
  phone: string
  addressLine1: string
  addressLine2: string
  city: string
  state: string
  postalCode: string
  country: string
  emailNotifications: boolean
  marketingEmails: boolean
  defaultRoomType: string
  defaultStyle: string
}

export type ProfileUpdateRequest = {
  email?: string
  full_name?: string
  company_name?: string
  phone?: string
  billing_address?: {
    line1?: string
    line2?: string
    city?: string
    state?: string
    postal_code?: string
    country?: string
  }
  profile_photo_url?: string
  preferences?: {
    email_notifications?: boolean
    marketing_emails?: boolean
    default_room_type?: string
    default_style?: string
  }
}

export function toFormData(profile: BackendProfile): ProfileFormData {
  const addr = profile.billing_address || {}
  const prefs = profile.preferences || {}
  return {
    fullName: profile.full_name || '',
    companyName: profile.company_name || '',
    phone: profile.phone || '',
    addressLine1: addr.line1 || '',
    addressLine2: addr.line2 || '',
    city: addr.city || '',
    state: addr.state || '',
    postalCode: addr.postal_code || '',
    country: addr.country || 'US',
    emailNotifications: prefs.email_notifications ?? true,
    marketingEmails: prefs.marketing_emails ?? false,
    defaultRoomType: prefs.default_room_type || 'living_room',
    defaultStyle: prefs.default_style || 'modern',
  }
}

export function buildUpdatePayload(form: ProfileFormData): ProfileUpdateRequest {
  return {
    full_name: form.fullName || undefined,
    company_name: form.companyName || undefined,
    phone: form.phone || undefined,
    billing_address: {
      line1: form.addressLine1 || undefined,
      line2: form.addressLine2 || undefined,
      city: form.city || undefined,
      state: form.state || undefined,
      postal_code: form.postalCode || undefined,
      country: form.country || undefined,
    },
    preferences: {
      email_notifications: form.emailNotifications,
      marketing_emails: form.marketingEmails,
      default_room_type: form.defaultRoomType || undefined,
      default_style: form.defaultStyle || undefined,
    },
  }
}
