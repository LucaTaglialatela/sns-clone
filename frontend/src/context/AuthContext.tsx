import { createContext, useState, useContext, useEffect, type ReactNode } from 'react';

interface AuthContextType {
  isAuthenticated: boolean;
  userId: string;
  userName: string;
  isLoading: boolean;
  following: string[];
  followUser: (targetUserId: string) => Promise<void>;
  unfollowUser: (targetUserId: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [userId, setUserId] = useState<string>("");
  const [userName, setUserName] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);
  const [following, setFollowing] = useState<string[]>([]);

  useEffect(() => {
    const checkAuthStatus = async () => {
      try {
        const response = await fetch('/me');
        if (response.ok) {
            const data = await response.json();
            setUserId(data.id);
            setUserName(data.name)
            setFollowing(data.following || []);
        } else {
            setUserId("");
            setUserName("");
            setFollowing([]);
        }
      } catch (error) {
        console.error('Could not fetch auth status:', error);
        setUserId("");
        setUserName("");
        setFollowing([]);
      } finally {
        setIsLoading(false);
      }
    };
    checkAuthStatus();
  }, []);

  const followUser = async (targetUserId: string) => {
    // Optimistic Update: Update the UI immediately.
    setFollowing(currentFollowing => [...currentFollowing, targetUserId]);

    try {
      const response = await fetch('/users/follow', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ following_id: targetUserId }),
      });
      if (!response.ok) throw new Error('API call to follow failed');
    } catch (error) {
      console.error("Failed to follow user:", error);
      // If the API fails, revert the UI change.
      setFollowing(currentFollowing => currentFollowing.filter(id => id !== targetUserId));
      alert('Failed to follow user. Please try again.');
    }
  };
  
  const unfollowUser = async (targetUserId: string) => {
    // Optimistic Update: Update the UI immediately.
    setFollowing(currentFollowing => currentFollowing.filter(id => id !== targetUserId));

    try {
      const response = await fetch('/users/unfollow', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ unfollowing_id: targetUserId }),
      });
       if (!response.ok) throw new Error('API call to unfollow failed');
    } catch (error) {
      console.error("Failed to unfollow user:", error);
      // If the API fails, revert the UI change.
      setFollowing(currentFollowing => [...currentFollowing, targetUserId]);
      alert('Failed to unfollow user. Please try again.');
    }
  };

  const contextValue = {
    isAuthenticated: !!userId,
    userId,
    userName,
    isLoading,
    following,
    followUser,
    unfollowUser,
  };

  return (
    <AuthContext.Provider value={contextValue}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
