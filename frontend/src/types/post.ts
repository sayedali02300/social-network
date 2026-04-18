export interface User {
    id: string;
    firstName: string;
    lastName: string;
    nickname: string;
    avatar: string;
}

export interface Post {
    id: string;
    userId: string;
    groupId?: string;
    title: string;
    content: string;
    imagePath: string;
    privacy: 'public' | 'almost_private' | 'private';
    createdAt: string;
    author: User;
    likeCount?: number;
    dislikeCount?: number;
    myReaction?: number;
}
