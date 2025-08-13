import { NextApiRequest, NextApiResponse } from 'next';
import { type NextRequest, NextResponse } from 'next/server';
import { revalidateAll } from '@/app/data';

const API_TOKEN = process.env.AREWETURBOYET_TOKEN;

interface RevalidationSuccess {
  revalidated: true;
}

interface RevalidationError {
  error?: string;
}

type Revalidation = RevalidationSuccess | RevalidationError;

// Revalidates all of the data caches associated with this deployment. Intended
// to be called from GitHub actions after new data is pushed to the KV store, so
// it can be reflected immediately in the UI.
//
// Note: areweturboyet and arewerspackyet must be revalidated independently, as
// they're separate vercel projects with separate data caches.
//
// Example: https://nextjs.org/docs/app/api-reference/functions/revalidateTag#route-handler
export async function POST(
  req: NextRequest,
): Promise<NextResponse<Revalidation>> {
  // Check for the API key in the request headers. This isn't particularly
  // sensitive, but it could cost us money if somebody hit it maliciously.
  const headerToken = req.headers.get('X-Auth-Token');
  if (!API_TOKEN || headerToken !== API_TOKEN) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
  }

  try {
    revalidateAll();
    return NextResponse.json({
      revalidated: true,
    });
  } catch (error) {
    return NextResponse.json(
      {
        error: (error as Error).message,
      },
      { status: 500 },
    );
  }
}
