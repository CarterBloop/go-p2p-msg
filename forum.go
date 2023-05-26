package main

import (
	"sync"
	"bufio"
	"fmt"
	"os"
	"strings"
	"os/exec"
	"runtime"
)

type Comment struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type Post struct {
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Content  string    `json:"content"`
	Comments []Comment `json:"comments"`
	mutex    sync.RWMutex
}

type Room struct {
	Name  string  `json:"name"`
	Posts []Post  `json:"posts"`
	mutex sync.RWMutex
}

type Forum struct {
	Rooms map[string]*Room  `json:"rooms"`
	mutex sync.RWMutex
}

func (f *Forum) runForum() {
	reader := bufio.NewReader(os.Stdin)

	for {
		
		fmt.Println("Enter command (create-room, post, comment, list-rooms, list-posts, list-comments):")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		switch command {
		case "create-room":
			// Create a new room
			clearScreen()
			fmt.Println("Enter name for the new room:")
			roomName, _ := reader.ReadString('\n')
			roomName = strings.TrimSpace(roomName)
			room := &Room{Name: roomName, Posts: []Post{}}
			f.mutex.Lock()
			f.Rooms[roomName] = room
			f.mutex.Unlock()
			fmt.Println("Created room:", roomName)

		case "post":
			// Create a new post
			clearScreen()
			fmt.Println("Enter room name for the post:")
			roomName, _ := reader.ReadString('\n')
			roomName = strings.TrimSpace(roomName)

			f.mutex.RLock()
			room, ok := f.Rooms[roomName]
			f.mutex.RUnlock()
			if !ok {
				fmt.Println("Room not found:", roomName)
				break
			}

			fmt.Println("Enter post title:")
			postTitle, _ := reader.ReadString('\n')
			postTitle = strings.TrimSpace(postTitle)

			fmt.Println("Enter post content:")
			postContent, _ := reader.ReadString('\n')
			postContent = strings.TrimSpace(postContent)

			fmt.Println("Enter post author:")
			postAuthor, _ := reader.ReadString('\n')
			postAuthor = strings.TrimSpace(postAuthor)

			post := Post{Title: postTitle, Content: postContent, Author: postAuthor, Comments: []Comment{}}
			room.mutex.Lock()
			room.Posts = append(room.Posts, post)
			room.mutex.Unlock()
			fmt.Println("Created post:", postTitle)

		case "comment":
			// Handle the case for adding a comment here, similar to the "post" case.
			clearScreen()
			fmt.Println("Enter room name for the comment:")
			roomName, _ := reader.ReadString('\n')
			roomName = strings.TrimSpace(roomName)

			f.mutex.RLock()
			room, ok := f.Rooms[roomName]
			f.mutex.RUnlock()
			if !ok {
				fmt.Println("Room not found:", roomName)
				break
			}

			fmt.Println("Enter post title for the comment:")
			postTitle, _ := reader.ReadString('\n')
			postTitle = strings.TrimSpace(postTitle)

			var post *Post
			room.mutex.RLock()
			for i := range room.Posts {
				if room.Posts[i].Title == postTitle {
					post = &room.Posts[i]
					break
				}
			}
			room.mutex.RUnlock()

			if post == nil {
				fmt.Println("Post not found:", postTitle)
				break
			}

			fmt.Println("Enter comment content:")
			commentContent, _ := reader.ReadString('\n')
			commentContent = strings.TrimSpace(commentContent)

			fmt.Println("Enter comment author:")
			commentAuthor, _ := reader.ReadString('\n')
			commentAuthor = strings.TrimSpace(commentAuthor)

			comment := Comment{Content: commentContent, Author: commentAuthor}
			post.Comments = append(post.Comments, comment)
			fmt.Println("Added comment to post:", postTitle)

		case "list-rooms":
			// List all rooms
			clearScreen()
			f.mutex.RLock()
			for roomName := range f.Rooms {
				fmt.Println(roomName)
			}
			f.mutex.RUnlock()

		case "list-posts":
			// Ask for a room name and list all posts in that room
			clearScreen()
			fmt.Println("Enter room name:")
			roomName, _ := reader.ReadString('\n')
			roomName = strings.TrimSpace(roomName)

			f.mutex.RLock()
			room, ok := f.Rooms[roomName]
			f.mutex.RUnlock()
			if !ok {
				fmt.Println("Room not found:", roomName)
				break
			}

			room.mutex.RLock()
			for _, post := range room.Posts {
				fmt.Println(post.Title)
			}
			room.mutex.RUnlock()

		case "list-comments":
			// Ask for a room name and a post title, then list all comments under that post
			clearScreen()
			fmt.Println("Enter room name:")
			roomName, _ := reader.ReadString('\n')
			roomName = strings.TrimSpace(roomName)

			f.mutex.RLock()
			room, ok := f.Rooms[roomName]
			f.mutex.RUnlock()
			if !ok {
				fmt.Println("Room not found:", roomName)
				break
			}

			fmt.Println("Enter post title:")
			postTitle, _ := reader.ReadString('\n')
			postTitle = strings.TrimSpace(postTitle)

			var post *Post
			room.mutex.RLock()
			for i := range room.Posts {
				if room.Posts[i].Title == postTitle {
					post = &room.Posts[i]
					break
				}
			}
			room.mutex.RUnlock()

			if post == nil {
				fmt.Println("Post not found:", postTitle)
				break
			}

			for _, comment := range post.Comments {
				fmt.Println(comment.Content)
			}

		default:
			clearScreen()
			fmt.Println("Unknown command:", command)
		}
	}
}

func newForum() *Forum {
	return &Forum{
		Rooms: make(map[string]*Room),
		mutex: sync.RWMutex{},
	}
}

func clearScreen() {
	var clearCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		clearCmd = exec.Command("cmd", "/c", "cls")
	} else {
		clearCmd = exec.Command("clear")
	}
	clearCmd.Stdout = os.Stdout
	clearCmd.Run()
}

func main() {
	forum := newForum()
	forum.runForum()
}

