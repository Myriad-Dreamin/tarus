package oci_judge

import (
	context "context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	hr_bytes "github.com/Myriad-Dreamin/tarus/pkg/hr-bytes"
	tarus_io "github.com/Myriad-Dreamin/tarus/pkg/tarus-io"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	tarus_store "github.com/Myriad-Dreamin/tarus/pkg/tarus-store"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	v1 "github.com/containerd/containerd/metrics/types/v1"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/typeurl"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type ContainerdJudgeServiceServer struct {
	tarus.UnimplementedJudgeServiceServer
	client       *containerd.Client
	sessionStore tarus_store.JudgeSessionStore
	closers      []io.Closer
	ioRouter     tarus_io.Router
}

func NewContainerdServer() (svc *ContainerdJudgeServiceServer, err error) {
	svc = &ContainerdJudgeServiceServer{
		ioRouter: tarus_io.Statics,
	}

	svc.client, err = containerd.New("/run/containerd/containerd.sock",
		containerd.WithDefaultNamespace("tarus"))
	if err != nil {
		return
	}

	svc.closers = append(svc.closers, svc.client)
	defer func() {
		if err != nil {
			_ = svc.Close()
		}
	}()

	b, err := bbolt.Open("./test.db", os.FileMode(0644), nil)
	if err != nil {
		return
	}
	svc.sessionStore = tarus_store.NewJudgeSessionStore(tarus_store.NewDB(b))
	if svc.sessionStore == nil {
		err = errors.Wrapf(errdefs.ErrInvalidArgument, "session store not filled")
		return
	}

	if c, ok := svc.sessionStore.(io.Closer); ok {
		svc.closers = append(svc.closers, c)
	}
	return
}

func (c *ContainerdJudgeServiceServer) Close() error {
	for i := range c.closers {
		_ = c.closers[i].Close()
	}
	return nil
}

func (c *ContainerdJudgeServiceServer) Handshake(ctx context.Context, request *tarus.HandshakeRequest) (*tarus.HandshakeResponse, error) {
	return c.UnimplementedJudgeServiceServer.Handshake(ctx, request)
}

func (c *ContainerdJudgeServiceServer) CopyFile(ctx context.Context, request *tarus.CopyRequest) (*emptypb.Empty, error) {
	return c.UnimplementedJudgeServiceServer.CopyFile(ctx, request)
}

func (c *ContainerdJudgeServiceServer) CreateContainer(ctx context.Context, request *tarus.CreateContainerRequest) (_ *emptypb.Empty, err error) {
	ctx = namespaces.WithNamespace(ctx, "tarus")
	snapshotter := containerd.DefaultSnapshotter
	if err = c.prepareImageOnSnapshotter(ctx, request.ImageId, snapshotter); err != nil {
		return nil, err
	}

	fixedContainerId := fmt.Sprintf("tarus-engine-snapshot%d", 0)
	fixedContainerSnapshotId := fmt.Sprintf("tarus-engine-snapshot%d", 0)

	if err = c.client.SnapshotService(snapshotter).Remove(ctx, fixedContainerSnapshotId); err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	}
	if err = c.client.ContainerService().Delete(ctx, fixedContainerId); err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	}

	image, err := c.client.GetImage(ctx, request.ImageId)
	if err != nil {
		return nil, err
	}
	mounts := make([]specs.Mount, 0)

	m, err := filepath.Abs("./data/workdir-judge-engine0")
	if err != nil {
		return nil, err
	}
	mounts = append(mounts, specs.Mount{
		// notes: linux only flags
		Type:        "bind",
		Source:      m,
		Destination: "/workdir",
		Options:     []string{"rbind", "ro"},
	})

	container, err := c.client.NewContainer(
		ctx,
		fixedContainerId,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(fixedContainerSnapshotId, image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithMounts(mounts),
			oci.WithProcessCwd("/workdir"),
		),
		// containerd.WithRuntime("/home/kamiyoru/work/go/tarus/cmd/runsc/wrapper.template.sh", nil),
	)

	if err != nil {
		return nil, err
	}
	if container == nil {
		return nil, status.Error(codes.Internal, "container not created when successfully returning")
	}

	var session = tarus.OCIJudgeSession{
		CommitStatus: 0,
		ContainerId:  fixedContainerId,
		BinTarget:    request.GetBinTarget(),
		HostWorkdir:  "/workdir",
	}
	err = c.sessionStore.SetJudgeSession(ctx, request.TaskKey, &session)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *ContainerdJudgeServiceServer) RemoveContainer(ctx context.Context, request *tarus.RemoveContainerRequest) (*emptypb.Empty, error) {
	ctx = namespaces.WithNamespace(ctx, "tarus")
	fixedContainerId := fmt.Sprintf("tarus-engine-snapshot%d", 0)
	cc, err := c.client.LoadContainer(ctx, fixedContainerId)
	if err != nil {
		return nil, err
	}

	err = c.sessionStore.FinishSession(ctx, request.TaskKey, func() error {
		return cc.Delete(ctx, containerd.WithSnapshotCleanup)
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *ContainerdJudgeServiceServer) MakeJudge(rawCtx context.Context, request *tarus.MakeJudgeRequest) (*tarus.MakeJudgeResponse, error) {
	ctx := namespaces.WithNamespace(rawCtx, "tarus")

	fixedContainerId := "tarus-engine-snapshot0"
	cc, err := c.client.LoadContainer(ctx, fixedContainerId)
	if err != nil {
		return nil, err
	}

	var resp = new(tarus.MakeJudgeResponse)
	for i := range request.Testcases {
		err = c.withFreshTask(ctx, cc, func(t containerd.Task) error {
			// fmt.Printf("linux container create successfully\n")
			s, err := cc.Spec(ctx)
			if err != nil {
				return err
			}
			procTmpl := *s.Process

			session, err := c.sessionStore.GetJudgeSession(ctx, request.TaskKey)
			if err != nil {
				return err
			}

			var judgePoint = request.Testcases[i]
			procOpts := procTmpl
			procOpts.Terminal = false
			procOpts.Args = []string{session.BinTarget}

			var ioProvider = judgePoint.IoProvider
			if len(ioProvider) == 0 {
				ioProvider = request.IoProvider
			}

			ioc, err := c.ioRouter.MakeIOChannel(ioProvider)
			if err != nil {
				return err
			}

			fac, err := ioc(judgePoint.Input, judgePoint.Answer)
			if err != nil {
				return err
			}

			err = (func() error {
				process, err := t.Exec(ctx, "judge_exec", &procOpts, fac.AsCreator())
				if err != nil {
					return err
				}
				defer func() {
					_, _ = process.Delete(ctx)
				}()

				statusC, err := process.Wait(ctx)
				if err != nil {
					return err
				}

				var processStartTime = time.Now()
				if err := process.Start(ctx); err != nil {
					return err
				}
				defer func() {
					// todo: check process status
					_ = c.killProcess(ctx, process)
				}()

				select {
				case st := <-statusC:
					var qr = &tarus.QueryJudgeItem{
						JudgeKey: judgePoint.JudgeKey,
					}
					var jh JudgeHint
					code, exitedAt, err := st.Result()
					if err != nil {
						return err
					}
					jh.Code = int(code)
					qr.TimeUseHard = int64(exitedAt.Sub(processStartTime) * time.Nanosecond)

					s, err := fac.GetJudgeResult()
					if err != nil {
						return err
					}
					jh.CheckerResult = string(s)

					m, err := t.Metrics(rawCtx)
					if err != nil {
						return err
					}

					m0, err := typeurl.UnmarshalAny(m.Data)
					if err != nil {
						return err
					}
					if m2, ok := m0.(*v1.Metrics); ok {
						//fmt.Printf("metrics cpu result: %v %v %v\n", m2.CPU.Usage.Total, m2.CPU.Usage.Kernel, m2.CPU.Usage.User)
						//fmt.Printf("metrics memory result: %v %v %v\n", m2.Memory.Usage.Max, m2.Memory.Usage.Usage, m2.Memory.RSS)
						qr.TimeUse = int64(m2.CPU.Usage.User)
						qr.MemoryUse = int64(m2.Memory.Usage.Max)
					} else {
						fmt.Println("invalid type url for extracting metrics", m.Data.TypeUrl)
					}

					qr.Hint, err = json.Marshal(jh)
					if err != nil {
						return err
					}

					qr.Status, err = fac.GetJudgeStatus(s)
					if err != nil {
						return err
					}
					resp.Items = append(resp.Items, qr)
				case <-time.After(time.Second * 3):
					fmt.Printf("linux container timeout stop\n")
				}

				return nil
			})()
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (c *ContainerdJudgeServiceServer) QueryJudge(ctx context.Context, request *tarus.QueryJudgeRequest) (*tarus.QueryJudgeResponse, error) {
	return c.UnimplementedJudgeServiceServer.QueryJudge(ctx, request)
}

func (c *ContainerdJudgeServiceServer) ImportOCIArchive(ctx context.Context, fp string) error {
	ctx = namespaces.WithNamespace(ctx, "tarus")

	f, err := os.OpenFile(fp, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	_, err = c.client.Import(ctx, f)
	_ = f.Close()
	if err != nil {
		return err
	}

	return nil
}

var tk = []byte("transient:")

func (c *ContainerdJudgeServiceServer) TransientJudge(rawCtx context.Context, req *tarus_judge.TransientJudgeRequest) error {
	if req.TaskKey == nil {
		token := make([]byte, 12)
		_, err := rand.Read(token)
		if err != nil {
			return err
		}
		key := make([]byte, 24+len(tk))
		copy(key[:len(tk)], tk)
		copy(key[len(tk):], hex.EncodeToString(token))
		req.TaskKey = key
	}

	return tarus_judge.WithContainerEnvironment(c, rawCtx, req, func(rawCtx context.Context, req *tarus_judge.TransientJudgeRequest) error {
		req.IsAsync = false
		resp, err := c.MakeJudge(rawCtx, req.MakeJudgeRequest)
		if err != nil {
			return err
		}

		for i := range resp.Items {
			fmt.Printf("req %d judge: %v/%v/%v, resp: %v\n",
				i,
				time.Duration(resp.Items[i].TimeUse)*time.Nanosecond,
				time.Duration(resp.Items[i].TimeUseHard)*time.Nanosecond,
				hr_bytes.Byte(resp.Items[i].MemoryUse), resp.Items[i])
		}
		return nil
	})
}
