export type FileInfo = {
    id: number
    name: string
    size: number
    contentType: string
    createdAt: string
    updatedAt: string
}

export type Room = {
    code: string
    name: string
    expiresAt: string
    files?: FileInfo[]
}