import { describe, it, expect } from 'vitest'
import { toFormData, buildUpdatePayload, type BackendProfile, type ProfileFormData } from './profile'

describe('profile mappings', () => {
  it('success: toFormData maps full backend profile to camelCase form data', () => {
    const backend: BackendProfile = {
      id: 'user-1',
      email: 'user@example.com',
      full_name: 'John Doe',
      company_name: 'Acme Inc',
      phone: '+123456789',
      billing_address: {
        line1: '123 Main St',
        line2: 'Apt 4',
        city: 'Metropolis',
        state: 'CA',
        postal_code: '94102',
        country: 'US',
      },
      profile_photo_url: 'https://example.com/p.jpg',
      preferences: {
        email_notifications: true,
        marketing_emails: false,
        default_room_type: 'living_room',
        default_style: 'modern',
      },
      role: 'user',
      stripe_customer_id: 'cus_123',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }

    const form = toFormData(backend)

    expect(form).toEqual<ProfileFormData>({
      fullName: 'John Doe',
      companyName: 'Acme Inc',
      phone: '+123456789',
      addressLine1: '123 Main St',
      addressLine2: 'Apt 4',
      city: 'Metropolis',
      state: 'CA',
      postalCode: '94102',
      country: 'US',
      emailNotifications: true,
      marketingEmails: false,
      defaultRoomType: 'living_room',
      defaultStyle: 'modern',
    })
  })

  it('success: toFormData provides sensible defaults when fields are missing', () => {
    const backend: BackendProfile = {
      id: 'user-2',
      role: 'user',
      created_at: '',
      updated_at: '',
    }

    const form = toFormData(backend)

    expect(form).toEqual<ProfileFormData>({
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
    })
  })

  it('success: buildUpdatePayload maps form data to snake_case backend payload', () => {
    const form: ProfileFormData = {
      fullName: 'Jane Smith',
      companyName: 'Widgets Co',
      phone: '+1987654321',
      addressLine1: '456 Oak Ave',
      addressLine2: '',
      city: 'Gotham',
      state: 'NY',
      postalCode: '10001',
      country: 'US',
      emailNotifications: false,
      marketingEmails: true,
      defaultRoomType: 'bedroom',
      defaultStyle: 'contemporary',
    }

    const payload = buildUpdatePayload(form)

    expect(payload).toEqual({
      full_name: 'Jane Smith',
      company_name: 'Widgets Co',
      phone: '+1987654321',
      billing_address: {
        line1: '456 Oak Ave',
        line2: undefined,
        city: 'Gotham',
        state: 'NY',
        postal_code: '10001',
        country: 'US',
      },
      preferences: {
        email_notifications: false,
        marketing_emails: true,
        default_room_type: 'bedroom',
        default_style: 'contemporary',
      },
    })

    // Ensure JSON serialization drops undefined keys gracefully
    const json = JSON.stringify(payload)
    expect(json).toContain('"line1":"456 Oak Ave"')
    expect(json).not.toContain('line2')
  })
})
