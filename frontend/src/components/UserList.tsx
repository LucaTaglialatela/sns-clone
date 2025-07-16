import React, { useState, useEffect } from "react";
import type { User } from "../types/User";
import { useAuth } from "../context/AuthContext";
import UserProfile from "./UserProfile";

interface UserListProps {
  selectedUser: User | null;
  setSelectedUser: (user: User | null) => void;
  setIsViewingProfile: (isViewingProfile: boolean) => void;
}

const UserList: React.FC<UserListProps> = ({
  selectedUser,
  setSelectedUser,
  setIsViewingProfile,
}) => {
  const {
    userId: currentUserId,
    following,
    followUser,
    unfollowUser,
  } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [isUsersLoading, setIsUsersLoading] = useState(true);
  const [usersError, setUsersError] = useState<string | null>(null);
  const [actionInProgress, setActionInProgress] = useState<string | null>(null);

  useEffect(() => {
    setIsViewingProfile(!!selectedUser);
  }, [selectedUser]);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await fetch("/users");
        if (!response.ok)
          throw new Error(`HTTP error! status: ${response.status}`);
        const data: User[] = await response.json();
        setUsers(data);
      } catch (e) {
        if (e instanceof Error) setUsersError(e.message);
        else setUsersError("An unknown error occurred");
      } finally {
        setIsUsersLoading(false);
      }
    };
    fetchUsers();
  }, []);

  const handleFollowToggle = async (
    targetUserId: string,
    isCurrentlyFollowing: boolean
  ) => {
    setActionInProgress(targetUserId);
    if (isCurrentlyFollowing) {
      await unfollowUser(targetUserId);
    } else {
      await followUser(targetUserId);
    }
    setActionInProgress(null);
  };

  if (isUsersLoading) {
    return (
      <div className="text-center text-gray-500 pt-10">Loading users...</div>
    );
  }

  if (usersError) {
    return (
      <div className="text-center text-red-500 pt-10">Error: {usersError}</div>
    );
  }

  if (selectedUser) {
    const isFollowing = following.includes(selectedUser.id);
    const isButtonLoading = actionInProgress === selectedUser.id;

    return (
      <UserProfile
        user={selectedUser}
        isFollowing={isFollowing}
        isButtonLoading={isButtonLoading}
        setSelectedUser={setSelectedUser}
        handleFollowToggle={handleFollowToggle}
        returnButtonText={"Back to all users"}
      />
    );
  }

  return (
    <div>
      <h2 className="text-2xl font-bold mb-4">Registered Users</h2>
      <ul className="space-y-3">
        {users.map((user) => {
          const isFollowing = following.includes(user.id);
          const isButtonLoading = actionInProgress === user.id;

          return (
            <li
              key={user.id}
              className="p-4 rounded-lg flex justify-between items-center bg-white shadow-lg"
            >
              <div className="flex items-center">
                <img
                  src={user.picture}
                  alt="user's picture"
                  className="max-w-full mr-6 h-[80px] rounded-full inset-shadow-sm"
                />
                <div>
                  <button
                    onClick={() => setSelectedUser(user)}
                    className="font-semibold text-gray-800 text-left cursor-pointer"
                  >
                    {user.name}
                  </button>
                  <p className="text-sm text-gray-600">{user.email}</p>
                </div>
              </div>
              {user.id !== currentUserId && (
                <button
                  onClick={() => handleFollowToggle(user.id, isFollowing)}
                  disabled={isButtonLoading}
                  className={`w-28 text-center px-4 py-1.5 cursor-pointer text-sm font-semibold rounded-full transition-colors duration-200 disabled:opacity-50 ${
                    isFollowing
                      ? "bg-white text-gray-700 border border-gray-300 hover:bg-gray-100"
                      : "bg-black text-white hover:bg-gray-800"
                  }`}
                >
                  {isButtonLoading
                    ? "..."
                    : isFollowing
                    ? "Following"
                    : "Follow"}
                </button>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
};

export default UserList;
