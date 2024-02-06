// Pages 155-156
// Listing 7-13: Retrieving the peer credentials for a socket connection.
package auth

import (
	"log"
	"net"
	"os/user"

	unix "golang.org/x/sys/unix"
)

func Allowed(conn *net.UnixConn, groups map[string]struct{}) bool {
	if conn == nil || groups == nil || len(groups) == 0 {
		return false
	}

	// To retrieve the peer's Unix credentials, we first grab the underlying
	// file object from net.UnixConn, the object that represents our side of the
	// Unix domain socket connection.
	// Since we need to extract the file descriptor details from the connection,
	// we cannot simply rely on the net.Conn interface that we receive from the
	// listener's Accept method.
	// Instead, our Allowed function requires the caller to pass in a pointer to
	// the underlying net.UnixConn object, typically returned from the
	// listener's AcceptUnix method.
	file, _ := conn.File()
	defer func() { _ = file.Close() }()

	var (
		err   error
		ucred *unix.Ucred
	)

	for {
		// We can then pass the file object's descriptor, the protocol-level
		// unix.SOL_SOCKET, and the option name unix.SO_PEERCRED to the
		// unix.GetsockoptUcred function.
		// Retrieving socket options from the Linux kernel requires that we
		// specify both the option we want adn the level at which the option
		// resides.
		// The unix.SOL_SOCKET tells the Linux kernel we want a socket-level
		// option, as opposed to, for example, unix.SOL_TCP, which indicates
		// TCP-level options.
		// The unix.SO_PEERCRED constant tells the Linux kernel that we want the
		// peer credentials option.
		// If the Linux kernel finds the peer credentials option at the Unix
		// domain socket level, unix.GetsockoptUcred returns a pointer to a
		// valid unix.Ucred object.
		// The unix.Ucred object contains the peer's process, user, and group IDs.
		ucred, err = unix.GetsockoptUcred(int(file.Fd()), unix.SOL_SOCKET, unix.SO_PEERCRED)
		if err == unix.EINTR {
			continue // syscall interrupted, try again
		}
		if err != nil {
			log.Println(err)
			return false
		}

		break
	}

	// We pass the peer's user ID to the user.LookupId function.
	u, err := user.LookupId(string(ucred.Uid))
	if err != nil {
		log.Println(err)
		return false
	}

	// If successful, we then retrieve a list of group IDs from the user object.
	gids, err := u.GroupIds()
	if err != nil {
		log.Println(err)
		return false
	}
	
	for _, gid := range gids {
		// The user can belong to more than one group, and we want to consider
		// each one for access.
		// We check each group ID against a map of allowed groups.
		// If any one of hte peer's group IDs is in our map, we return true,
		// allowing the peer to connect.
		if _, ok := groups[gid]; ok {
			return true
		}
	}

	return false
}
