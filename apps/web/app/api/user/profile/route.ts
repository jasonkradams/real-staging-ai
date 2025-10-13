import { NextResponse } from 'next/server';

/**
 * Mock in-memory storage for development
 * In production, this data will be stored in the database via backend API
 */
interface BillingAddress {
  line1?: string;
  line2?: string;
  city?: string;
  state?: string;
  postalCode?: string;
  country?: string;
}

let mockProfile = {
  id: 'mock-user-id',
  email: 'user@example.com',
  fullName: null as string | null,
  companyName: null as string | null,
  phone: null as string | null,
  billingAddress: null as BillingAddress | null,
  profilePhotoUrl: null as string | null,
  preferences: {
    emailNotifications: true,
    marketingEmails: false,
    defaultRoomType: 'living_room',
    defaultStyle: 'modern',
  },
};

/**
 * GET /api/user/profile
 * Returns the current user's profile information
 * 
 * TODO: This is a MOCK endpoint for development only.
 * Replace with actual backend API call once implemented.
 * Backend endpoint should be: GET /api/v1/user/profile
 */
export async function GET() {
  try {
    // Return the stored mock profile
    return NextResponse.json(mockProfile);
  } catch (error) {
    console.error('Error fetching profile:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}

/**
 * PATCH /api/user/profile
 * Updates the current user's profile information
 * 
 * TODO: This is a MOCK endpoint for development only.
 * Replace with actual backend API call once implemented.
 * Backend endpoint should be: PATCH /api/v1/user/profile
 */
export async function PATCH(request: Request) {
  try {
    const body = await request.json();

    // Update the mock profile storage
    mockProfile = {
      ...mockProfile,
      ...body,
    };

    // Return the updated profile
    return NextResponse.json(mockProfile);
  } catch (error) {
    console.error('Error updating profile:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
