'use client';

import dynamic from 'next/dynamic';

export const WebShell = dynamic(() => import('./web_shell'), {
  ssr: false,
  loading: () => <div>Loading terminal...</div>
});
