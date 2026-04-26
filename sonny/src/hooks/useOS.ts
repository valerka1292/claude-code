import React from 'react';

export function useOS() {
  const [os, setOS] = React.useState<'mac' | 'win' | 'linux'>('win');
  
  React.useEffect(() => {
    if (window.electron?.platform) {
      const platform = window.electron.platform;
      if (platform === 'darwin') setOS('mac');
      else if (platform === 'linux') setOS('linux');
      else setOS('win');
    } else {
      const platform = window.navigator.platform.toLowerCase();
      if (platform.includes('mac')) setOS('mac');
      else if (platform.includes('linux')) setOS('linux');
      else setOS('win');
    }
  }, []);
  
  return os;
}
